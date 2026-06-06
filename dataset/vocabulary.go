package dataset

import (
	"fmt"
	"math"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

type Vocabulary struct {
	Guesses          []words.Word
	Solutions        []words.Word
	guessIDs         map[string]uint16
	solutionIDs      map[string]uint16
	guessSolutionIDs []int
	solutionGuessIDs []uint16
}

func NewVocabulary(guesses []words.Word, solutions []words.Word) (*Vocabulary, error) {
	if len(guesses) > math.MaxUint16 {
		return nil, fmt.Errorf("guess vocabulary has %d words, maximum is %d", len(guesses), math.MaxUint16)
	}
	if len(solutions) > math.MaxUint16 {
		return nil, fmt.Errorf("solution vocabulary has %d words, maximum is %d", len(solutions), math.MaxUint16)
	}

	vocab := &Vocabulary{
		Guesses:          append([]words.Word(nil), guesses...),
		Solutions:        append([]words.Word(nil), solutions...),
		guessIDs:         make(map[string]uint16, len(guesses)),
		solutionIDs:      make(map[string]uint16, len(solutions)),
		guessSolutionIDs: make([]int, len(guesses)),
		solutionGuessIDs: make([]uint16, len(solutions)),
	}

	for i := range vocab.guessSolutionIDs {
		vocab.guessSolutionIDs[i] = -1
	}

	for i, guess := range guesses {
		word := string(guess)
		if _, exists := vocab.guessIDs[word]; exists {
			return nil, fmt.Errorf("duplicate guess word %q", word)
		}
		vocab.guessIDs[word] = uint16(i)
	}

	for i, solution := range solutions {
		word := string(solution)
		if _, exists := vocab.solutionIDs[word]; exists {
			return nil, fmt.Errorf("duplicate solution word %q", word)
		}
		guessID, exists := vocab.guessIDs[word]
		if !exists {
			return nil, fmt.Errorf("solution word %q is not in the guess vocabulary", word)
		}

		vocab.solutionIDs[word] = uint16(i)
		vocab.solutionGuessIDs[i] = guessID
		vocab.guessSolutionIDs[guessID] = i
	}

	return vocab, nil
}

func (v *Vocabulary) GuessID(word words.Word) (uint16, bool) {
	id, ok := v.guessIDs[string(word)]
	return id, ok
}

func (v *Vocabulary) SolutionID(word words.Word) (uint16, bool) {
	id, ok := v.solutionIDs[string(word)]
	return id, ok
}

func (v *Vocabulary) SolutionGuessID(solutionID uint16) uint16 {
	return v.solutionGuessIDs[solutionID]
}

func (v *Vocabulary) GuessSolutionID(guessID uint16) int {
	return v.guessSolutionIDs[guessID]
}
