package dataset

import (
	"testing"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

func TestTeacherRanksByWorstCaseSizeThenPotentialSolution(t *testing.T) {
	guesses := []words.Word{"AAAAA", "AAAAB", "BBBBB"}
	solutions := []words.Word{"AAAAA", "BBBBB"}

	vocab, err := NewVocabulary(guesses, solutions)
	if err != nil {
		t.Fatalf("new vocabulary: %v", err)
	}
	matrix, err := NewFeedbackMatrix(vocab)
	if err != nil {
		t.Fatalf("new feedback matrix: %v", err)
	}

	labels, err := NewTeacher(vocab, matrix).Rank([]uint16{0, 1}, 2)
	if err != nil {
		t.Fatalf("rank: %v", err)
	}

	if labels[0].GuessID != 0 {
		t.Fatalf("top guess id = %d, want 0", labels[0].GuessID)
	}
	if labels[1].GuessID != 2 {
		t.Fatalf("second guess id = %d, want 2", labels[1].GuessID)
	}
	if labels[0].WorstCaseSize != 1 {
		t.Fatalf("worst-case size = %d, want 1", labels[0].WorstCaseSize)
	}
	if labels[0].ReductionRatio != 0.5 {
		t.Fatalf("reduction ratio = %f, want 0.5", labels[0].ReductionRatio)
	}
}
