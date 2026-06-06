package dataset

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"sync"

	"github.com/sam-bee/wordle-ml_game-engine/words"
	wordlists "github.com/sam-bee/wordle-ml_wordlists"
)

type Config struct {
	OutputDir       string
	Seed            int64
	Workers         int
	TopK            int
	MaxDepth        int
	RecordsPerDepth int
	IncludeOpening  bool
	ProgressWriter  io.Writer
}

type Result struct {
	Splits []SplitResult
}

type SplitResult struct {
	Name          string
	BinaryPath    string
	MetadataPath  string
	RecordCount   int
	SolutionCount int
}

type generatedSolution struct {
	SolutionID uint16
	Records    []Record
	Err        error
}

func Generate(ctx context.Context, config Config) (Result, error) {
	if err := config.validate(); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(config.OutputDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", config.OutputDir, err)
	}

	validGuesses, err := words.GetValidGuesses()
	if err != nil {
		return Result{}, fmt.Errorf("load valid guesses: %w", err)
	}
	validSolutions, err := words.GetValidSolutions()
	if err != nil {
		return Result{}, fmt.Errorf("load valid solutions: %w", err)
	}

	vocab, err := NewVocabulary(validGuesses, validSolutions)
	if err != nil {
		return Result{}, err
	}

	progress(config, "loaded %d valid guesses and %d valid solutions\n", len(vocab.Guesses), len(vocab.Solutions))
	progress(config, "precomputing feedback matrix...\n")

	matrix, err := NewFeedbackMatrix(vocab)
	if err != nil {
		return Result{}, err
	}
	teacher := NewTeacher(vocab, matrix)

	allSolutions := allSolutionIDs(len(vocab.Solutions))
	openingLabels, err := teacher.Rank(allSolutions, config.TopK)
	if err != nil {
		return Result{}, fmt.Errorf("rank opening state: %w", err)
	}

	splits := splitSolutions(len(vocab.Solutions), config.Seed)
	splitIndex := make(map[uint16]SplitID, len(vocab.Solutions))
	for _, split := range splits {
		for _, solutionID := range split.SolutionIDs {
			splitIndex[solutionID] = split.ID
		}
	}

	progress(config, "generating records with %d workers and seed %d...\n", config.Workers, config.Seed)
	solutionRecords, err := generateSolutions(ctx, config, vocab, matrix, teacher)
	if err != nil {
		return Result{}, err
	}

	recordsBySplit := make(map[SplitID][]Record, len(splits))
	if config.IncludeOpening {
		openingRecord, err := NewRecord(PaddingSolutionID, nil, allSolutions, openingLabels, "opening")
		if err != nil {
			return Result{}, fmt.Errorf("build opening record: %w", err)
		}
		recordsBySplit[SplitTrain] = append(recordsBySplit[SplitTrain], openingRecord)
	}

	for _, solutionRecord := range solutionRecords {
		splitID := splitIndex[solutionRecord.SolutionID]
		recordsBySplit[splitID] = append(recordsBySplit[splitID], solutionRecord.Records...)
	}

	wordlistHash := hashWordlists()
	generatorCommit, workingTreeDirty := generatorGitState()
	result := Result{Splits: make([]SplitResult, 0, len(splits))}

	for _, split := range splits {
		sortRecords(recordsBySplit[split.ID])
		paths := splitPaths(config.OutputDir, split.ID)
		if err := WriteBinaryFile(paths.binary, split.ID, recordsBySplit[split.ID], len(vocab.Guesses), len(split.SolutionIDs), config); err != nil {
			return Result{}, err
		}

		metadata := NewMetadata(
			split,
			paths.binary,
			recordsBySplit[split.ID],
			vocab,
			config,
			wordlistHash,
			generatorCommit,
			workingTreeDirty,
		)
		if err := WriteMetadataFile(paths.metadata, metadata); err != nil {
			return Result{}, err
		}

		result.Splits = append(result.Splits, SplitResult{
			Name:          split.ID.String(),
			BinaryPath:    paths.binary,
			MetadataPath:  paths.metadata,
			RecordCount:   len(recordsBySplit[split.ID]),
			SolutionCount: len(split.SolutionIDs),
		})
	}

	return result, nil
}

func (config Config) validate() error {
	if config.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	if config.Workers < 1 {
		return fmt.Errorf("workers must be at least 1, got %d", config.Workers)
	}
	if config.TopK != FixedTopK {
		return fmt.Errorf("topK must be %d for the current binary record format, got %d", FixedTopK, config.TopK)
	}
	if config.MaxDepth != MaxDepth {
		return fmt.Errorf("max depth must be %d for the current binary record format, got %d", MaxDepth, config.MaxDepth)
	}
	if config.RecordsPerDepth < 1 {
		return fmt.Errorf("records per depth must be at least 1, got %d", config.RecordsPerDepth)
	}
	return nil
}

func generateSolutions(ctx context.Context, config Config, vocab *Vocabulary, matrix *FeedbackMatrix, teacher *Teacher) ([]generatedSolution, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan uint16)
	results := make(chan generatedSolution)

	var workers sync.WaitGroup
	for i := 0; i < config.Workers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for solutionID := range jobs {
				records, err := generateSolutionRecords(config, vocab, matrix, teacher, solutionID)
				results <- generatedSolution{SolutionID: solutionID, Records: records, Err: err}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for solutionID := range vocab.Solutions {
			select {
			case <-ctx.Done():
				return
			case jobs <- uint16(solutionID):
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	generated := make([]generatedSolution, 0, len(vocab.Solutions))
	var firstErr error
	for result := range results {
		if result.Err != nil {
			if firstErr == nil {
				firstErr = result.Err
				cancel()
			}
			continue
		}
		generated = append(generated, result)
		if len(generated)%100 == 0 {
			progress(config, "generated records for %d/%d solutions\n", len(generated), len(vocab.Solutions))
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sort.Slice(generated, func(i, j int) bool {
		return generated[i].SolutionID < generated[j].SolutionID
	})

	return generated, nil
}

func generateSolutionRecords(config Config, vocab *Vocabulary, matrix *FeedbackMatrix, teacher *Teacher, solutionID uint16) ([]Record, error) {
	generator := solutionGenerator{
		config:   config,
		vocab:    vocab,
		matrix:   matrix,
		teacher:  teacher,
		rng:      rand.New(rand.NewSource(solutionSeed(config.Seed, solutionID))),
		solution: solutionID,
		buckets:  make(map[int][]Record, MaxDepth),
		seen:     make(map[int]map[string]bool, MaxDepth),
	}

	if err := generator.addTeacherTrajectory(); err != nil {
		return nil, err
	}
	if err := generator.fillRandomHistories(); err != nil {
		return nil, err
	}

	records := make([]Record, 0, MaxDepth*config.RecordsPerDepth)
	for depth := 1; depth <= MaxDepth; depth++ {
		records = append(records, generator.buckets[depth]...)
	}

	sortRecords(records)
	return records, nil
}

type solutionGenerator struct {
	config   Config
	vocab    *Vocabulary
	matrix   *FeedbackMatrix
	teacher  *Teacher
	rng      *rand.Rand
	solution uint16
	buckets  map[int][]Record
	seen     map[int]map[string]bool
}

func (g *solutionGenerator) addTeacherTrajectory() error {
	var history []Turn
	shortlist := allSolutionIDs(len(g.vocab.Solutions))

	for depth := 0; depth < MaxDepth; depth++ {
		labels, err := g.teacher.Rank(shortlist, g.config.TopK)
		if err != nil {
			return fmt.Errorf("rank teacher trajectory for solution %d at depth %d: %w", g.solution, depth, err)
		}

		if depth > 0 {
			if err := g.addRecord(history, shortlist, labels, "teacher"); err != nil {
				return err
			}
		}

		guessID := labels[0].GuessID
		if guessID == g.vocab.SolutionGuessID(g.solution) {
			return nil
		}

		history = append(history, Turn{
			GuessID:      guessID,
			FeedbackCode: g.matrix.FeedbackCode(g.solution, guessID),
		})
		shortlist = g.matrix.Shortlist(history)
	}

	return nil
}

func (g *solutionGenerator) fillRandomHistories() error {
	for depth := 1; depth <= MaxDepth; depth++ {
		attempts := 0
		for len(g.buckets[depth]) < g.config.RecordsPerDepth {
			attempts++
			if attempts > 100000 {
				return fmt.Errorf("could not generate enough depth-%d histories for solution %d after %d attempts", depth, g.solution, attempts)
			}

			history := g.randomHistory(depth)
			if g.hasSeen(depth, history) {
				continue
			}

			shortlist := g.matrix.Shortlist(history)
			labels, err := g.teacher.Rank(shortlist, g.config.TopK)
			if err != nil {
				return fmt.Errorf("rank random history for solution %d at depth %d: %w", g.solution, depth, err)
			}
			if err := g.addRecord(history, shortlist, labels, "random"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *solutionGenerator) randomHistory(depth int) []Turn {
	history := make([]Turn, 0, depth)
	used := make(map[uint16]bool, depth+1)
	used[g.vocab.SolutionGuessID(g.solution)] = true

	for len(history) < depth {
		guessID := uint16(g.rng.Intn(len(g.vocab.Guesses)))
		if used[guessID] {
			continue
		}

		used[guessID] = true
		history = append(history, Turn{
			GuessID:      guessID,
			FeedbackCode: g.matrix.FeedbackCode(g.solution, guessID),
		})
	}

	return history
}

func (g *solutionGenerator) addRecord(history []Turn, shortlist []uint16, labels []Label, source string) error {
	depth := len(history)
	if len(g.buckets[depth]) >= g.config.RecordsPerDepth {
		return nil
	}
	if g.hasSeen(depth, history) {
		return nil
	}

	record, err := NewRecord(g.solution, history, shortlist, labels, source)
	if err != nil {
		return err
	}

	g.markSeen(depth, history)
	g.buckets[depth] = append(g.buckets[depth], record)
	return nil
}

func (g *solutionGenerator) hasSeen(depth int, history []Turn) bool {
	if g.seen[depth] == nil {
		return false
	}
	return g.seen[depth][historyKey(history)]
}

func (g *solutionGenerator) markSeen(depth int, history []Turn) {
	if g.seen[depth] == nil {
		g.seen[depth] = make(map[string]bool)
	}
	g.seen[depth][historyKey(history)] = true
}

func allSolutionIDs(solutionCount int) []uint16 {
	ids := make([]uint16, solutionCount)
	for i := range ids {
		ids[i] = uint16(i)
	}
	return ids
}

func sortRecords(records []Record) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].SolutionID == PaddingSolutionID && records[j].SolutionID != PaddingSolutionID {
			return true
		}
		if records[i].SolutionID != PaddingSolutionID && records[j].SolutionID == PaddingSolutionID {
			return false
		}
		if records[i].SolutionID != records[j].SolutionID {
			return records[i].SolutionID < records[j].SolutionID
		}
		if records[i].TurnDepth != records[j].TurnDepth {
			return records[i].TurnDepth < records[j].TurnDepth
		}
		return records[i].HistoryDeduplicationID < records[j].HistoryDeduplicationID
	})
}

func solutionSeed(seed int64, solutionID uint16) int64 {
	value := uint64(seed) ^ (uint64(solutionID) + 0x9e3779b97f4a7c15)
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	value ^= value >> 31
	return int64(value)
}

func hashWordlists() string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(wordlists.ValidGuessesCSV()))
	_, _ = hasher.Write([]byte("\n---solutions---\n"))
	_, _ = hasher.Write([]byte(wordlists.ValidSolutionsCSV()))
	return hex.EncodeToString(hasher.Sum(nil))
}

func progress(config Config, format string, args ...any) {
	if config.ProgressWriter == nil {
		return
	}
	_, _ = fmt.Fprintf(config.ProgressWriter, format, args...)
}
