package dataset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

func TestWriteHumanReadableFileConvertsBinaryDataset(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "sample.bin")
	outputDir := filepath.Join(dir, "human-readable")

	record, err := NewRecord(
		mustTestVocabulary(t),
		1,
		[]Turn{{GuessID: 2, FeedbackCode: 5}},
		[]uint16{1},
		[]Label{{GuessID: 3, ReductionRatio: 0.75, WorstCaseSize: 1}},
		"test",
	)
	if err != nil {
		t.Fatalf("new record: %v", err)
	}

	config := Config{TopK: FixedTopK, MaxDepth: MaxDepth}
	if err := WriteBinaryFile(binaryPath, SplitTrain, []Record{record}, 10, 1, config); err != nil {
		t.Fatalf("write binary file: %v", err)
	}

	outputPath, err := WriteHumanReadableFile(binaryPath, outputDir)
	if err != nil {
		t.Fatalf("write human-readable file: %v", err)
	}
	if outputPath != filepath.Join(outputDir, "sample.json") {
		t.Fatalf("output path = %q, want sample.json in output dir", outputPath)
	}
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read human-readable output: %v", err)
	}
	var outputData map[string]any
	if err := json.Unmarshal(output, &outputData); err != nil {
		t.Fatalf("unmarshal human-readable output: %v", err)
	}

	readable, err := ReadHumanReadableDataset(binaryPath)
	if err != nil {
		t.Fatalf("read human-readable dataset: %v", err)
	}

	if readable.Header.RecordCount != 1 {
		t.Fatalf("record count = %d, want 1", readable.Header.RecordCount)
	}
	if len(readable.Records) != 1 {
		t.Fatalf("decoded records = %d, want 1", len(readable.Records))
	}
	got := readable.Records[0]
	if got.SolutionWord != "BBBBB" {
		t.Fatalf("solution word = %q, want BBBBB", got.SolutionWord)
	}
	if len(got.PreviousTurns) != 1 {
		t.Fatalf("previous turns = %d, want 1", len(got.PreviousTurns))
	}
	if got.PreviousTurns[0].FeedbackString != "GY---" {
		t.Fatalf("feedback string = %q, want GY---", got.PreviousTurns[0].FeedbackString)
	}
	if got.PreviousTurns[0].Feedback[0] != int(FeedbackGreen) || got.PreviousTurns[0].Feedback[1] != int(FeedbackYellow) {
		t.Fatalf("feedback values = %v, want numeric feedback values", got.PreviousTurns[0].Feedback)
	}
	if got.PreviousTurns[0].GuessWord != "CCCCC" {
		t.Fatalf("previous guess word = %q, want CCCCC", got.PreviousTurns[0].GuessWord)
	}
	outputRecord := outputData["records"].([]any)[0].(map[string]any)
	outputTurn := outputRecord["previous_turns"].([]any)[0].(map[string]any)
	if _, ok := outputTurn["feedback"].([]any); !ok {
		t.Fatalf("json feedback field = %T, want array", outputTurn["feedback"])
	}
	if got.TeacherRankedGuesses[0].GuessWord != "DDDDD" {
		t.Fatalf("top ranked guess = %q, want DDDDD", got.TeacherRankedGuesses[0].GuessWord)
	}
}

func mustTestVocabulary(t *testing.T) *Vocabulary {
	t.Helper()
	vocab, err := NewVocabulary(
		[]words.Word{"AAAAA", "BBBBB", "CCCCC", "DDDDD"},
		[]words.Word{"AAAAA", "BBBBB"},
	)
	if err != nil {
		t.Fatalf("new vocabulary: %v", err)
	}
	return vocab
}
