package dataset

import (
	"testing"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

func TestLoadVocabularyUsesActionSpaceForGuesses(t *testing.T) {
	vocab, err := loadVocabulary()
	if err != nil {
		t.Fatalf("load vocabulary: %v", err)
	}

	actionSpace, err := words.GetActionSpace()
	if err != nil {
		t.Fatalf("load action space: %v", err)
	}
	validGuesses, err := words.GetValidGuesses()
	if err != nil {
		t.Fatalf("load valid guesses: %v", err)
	}

	if len(vocab.Guesses) != len(actionSpace) {
		t.Fatalf("guess vocabulary length = %d, want action space length %d", len(vocab.Guesses), len(actionSpace))
	}
	if len(vocab.Guesses) == len(validGuesses) {
		t.Fatalf("guess vocabulary length = %d, still matches valid guesses", len(vocab.Guesses))
	}
	if vocab.Guesses[0] != actionSpace[0] {
		t.Fatalf("first guess = %q, want action space first guess %q", vocab.Guesses[0], actionSpace[0])
	}
}
