package dataset

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type outputPaths struct {
	binary   string
	metadata string
}

func splitPaths(outputDir string, splitID SplitID) outputPaths {
	stem := splitID.FileStem()
	return outputPaths{
		binary:   filepath.Join(outputDir, stem+".bin"),
		metadata: filepath.Join(outputDir, stem+".json"),
	}
}

func WriteBinaryFile(path string, splitID SplitID, records []Record, guessVocabSize int, solutionCount int, config Config) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create binary file %q: %w", path, err)
	}
	defer file.Close()

	if err := writeHeader(file, splitID, len(records), guessVocabSize, solutionCount, config); err != nil {
		return err
	}
	for i := range records {
		if err := writeRecord(file, records[i]); err != nil {
			return fmt.Errorf("write record %d to %q: %w", i, path, err)
		}
	}

	return nil
}

func writeHeader(file *os.File, splitID SplitID, recordCount int, guessVocabSize int, solutionCount int, config Config) error {
	if _, err := file.Write([]byte(FormatMagic)); err != nil {
		return err
	}

	fields := []uint32{
		FormatVersion,
		uint32(recordCount),
		uint32(config.TopK),
		uint32(config.MaxDepth),
		uint32(guessVocabSize),
		uint32(solutionCount),
		uint32(splitID),
	}

	for _, field := range fields {
		if err := binary.Write(file, binary.LittleEndian, field); err != nil {
			return err
		}
	}

	reserved := make([]byte, HeaderSizeBytes-4-len(fields)*4)
	_, err := file.Write(reserved)
	return err
}

func writeRecord(file *os.File, record Record) error {
	if _, err := file.Write(record.SolutionWord[:]); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, record.TurnDepth); err != nil {
		return err
	}
	for _, guessWord := range record.PreviousGuessWords {
		if _, err := file.Write(guessWord[:]); err != nil {
			return err
		}
	}
	for _, turnFeedback := range record.PreviousFeedback {
		for _, value := range turnFeedback {
			if err := binary.Write(file, binary.LittleEndian, value); err != nil {
				return err
			}
		}
	}
	if err := binary.Write(file, binary.LittleEndian, record.ShortlistSizeBefore); err != nil {
		return err
	}
	for _, guessWord := range record.TopKGuessWords {
		if _, err := file.Write(guessWord[:]); err != nil {
			return err
		}
	}
	for _, ratio := range record.TopKReductionRatios {
		if err := binary.Write(file, binary.LittleEndian, ratio); err != nil {
			return err
		}
	}
	for _, size := range record.TopKWorstCaseSizes {
		if err := binary.Write(file, binary.LittleEndian, size); err != nil {
			return err
		}
	}
	return nil
}

type Metadata struct {
	Version                   int      `json:"version"`
	Split                     string   `json:"split"`
	SplitID                   SplitID  `json:"split_id"`
	BinaryFile                string   `json:"binary_file"`
	RecordCount               int      `json:"record_count"`
	HeaderSizeBytes           int      `json:"header_size_bytes"`
	RecordSizeBytes           int      `json:"record_size_bytes"`
	TopK                      int      `json:"top_k"`
	MaxTurns                  int      `json:"max_turns"`
	GuessVocabSize            int      `json:"guess_vocab_size"`
	GlobalSolutionVocabSize   int      `json:"global_solution_vocab_size"`
	SolutionCount             int      `json:"solution_count"`
	SolutionIDs               []uint16 `json:"solution_ids"`
	RecordsPerSolution        int      `json:"records_per_solution"`
	RecordsPerDepth           int      `json:"records_per_depth"`
	IncludesOpeningState      bool     `json:"includes_opening_state"`
	OpeningSolutionWord       string   `json:"opening_solution_word"`
	PaddingWord               string   `json:"padding_word"`
	PaddingFeedbackValue      uint8    `json:"padding_feedback_value"`
	WordlistHash              string   `json:"wordlist_hash"`
	GeneratorCommit           string   `json:"generator_commit"`
	GeneratorWorkingTreeDirty bool     `json:"generator_working_tree_dirty"`
	GeneratedAtUTC            string   `json:"generated_at_utc"`
	Seed                      int64    `json:"seed"`
	TeacherName               string   `json:"teacher_name"`
	ScoreMeaning              string   `json:"score_meaning"`
	WordEncoding              string   `json:"word_encoding"`
	FeedbackConvention        string   `json:"feedback_convention"`
}

func NewMetadata(split Split, binaryPath string, records []Record, vocab *Vocabulary, config Config, wordlistHash string, generatorCommit string, workingTreeDirty bool) Metadata {
	solutionIDs := append([]uint16(nil), split.SolutionIDs...)
	sort.Slice(solutionIDs, func(i, j int) bool {
		return solutionIDs[i] < solutionIDs[j]
	})

	return Metadata{
		Version:                   FormatVersion,
		Split:                     split.ID.String(),
		SplitID:                   split.ID,
		BinaryFile:                filepath.Base(binaryPath),
		RecordCount:               len(records),
		HeaderSizeBytes:           HeaderSizeBytes,
		RecordSizeBytes:           RecordSizeBytes,
		TopK:                      config.TopK,
		MaxTurns:                  config.MaxDepth,
		GuessVocabSize:            len(vocab.Guesses),
		GlobalSolutionVocabSize:   len(vocab.Solutions),
		SolutionCount:             len(split.SolutionIDs),
		SolutionIDs:               solutionIDs,
		RecordsPerSolution:        config.RecordsPerDepth * config.MaxDepth,
		RecordsPerDepth:           config.RecordsPerDepth,
		IncludesOpeningState:      split.ID == SplitTrain && config.IncludeOpening,
		OpeningSolutionWord:       "",
		PaddingWord:               "",
		PaddingFeedbackValue:      PaddingFeedbackValue,
		WordlistHash:              wordlistHash,
		GeneratorCommit:           generatorCommit,
		GeneratorWorkingTreeDirty: workingTreeDirty,
		GeneratedAtUTC:            time.Now().UTC().Format(time.RFC3339),
		Seed:                      config.Seed,
		TeacherName:               "worst_case_shortlist_reduction",
		ScoreMeaning:              "Per-state worst-case shortlist reduction ratio. Higher is better. Not globally comparable across states.",
		WordEncoding:              "Word fields are fixed-width 5-byte uppercase ASCII strings; all-zero bytes pad unused word fields and the global opening-state solution.",
		FeedbackConvention:        "0 grey/absent, 1 yellow/present wrong position, 2 green/correct position, 255 pads unused feedback slots",
	}
}

func WriteMetadataFile(path string, metadata Metadata) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create metadata file %q: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("write metadata file %q: %w", path, err)
	}

	return nil
}
