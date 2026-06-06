package dataset

import (
	"testing"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

func TestNewRecordPadsUnusedHistoryAndLabels(t *testing.T) {
	vocab, err := NewVocabulary(
		[]words.Word{"AAAAA", "BBBBB", "CCCCC", "DDDDD", "EEEEE"},
		[]words.Word{"AAAAA", "BBBBB", "CCCCC"},
	)
	if err != nil {
		t.Fatalf("new vocabulary: %v", err)
	}

	record, err := NewRecord(
		vocab,
		2,
		[]Turn{{GuessID: 1, FeedbackCode: 5}},
		[]uint16{7, 8},
		[]Label{{GuessID: 3, ReductionRatio: 0.5, WorstCaseSize: 1}},
		"test",
	)
	if err != nil {
		t.Fatalf("new record: %v", err)
	}
	if wordBytesString(record.SolutionWord) != "CCCCC" {
		t.Fatalf("solution word = %q, want CCCCC", wordBytesString(record.SolutionWord))
	}

	if record.TurnDepth != 1 {
		t.Fatalf("turn depth = %d, want 1", record.TurnDepth)
	}
	if wordBytesString(record.PreviousGuessWords[0]) != "BBBBB" {
		t.Fatalf("first previous guess = %q, want BBBBB", wordBytesString(record.PreviousGuessWords[0]))
	}
	if wordBytesString(record.PreviousGuessWords[1]) != "" {
		t.Fatalf("unused previous guess = %q, want padding", wordBytesString(record.PreviousGuessWords[1]))
	}
	if record.PreviousFeedback[1][0] != PaddingFeedbackValue {
		t.Fatalf("unused feedback = %d, want padding", record.PreviousFeedback[1][0])
	}
	if wordBytesString(record.TopKGuessWords[0]) != "DDDDD" {
		t.Fatalf("first label guess = %q, want DDDDD", wordBytesString(record.TopKGuessWords[0]))
	}
	if wordBytesString(record.TopKGuessWords[1]) != "" {
		t.Fatalf("unused label guess = %q, want padding", wordBytesString(record.TopKGuessWords[1]))
	}
}
