package player

import (
	"testing"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

func TestChoosingGuess(t *testing.T) {

	table := []struct {
		name             string
		inputIsLastGuess bool
		inputActionSpace []words.Word
		inputSolutions   []words.Word
		expected         string
		verboseError     string
	}{
		{
			name:             "Basic",
			inputIsLastGuess: false,
			inputActionSpace: []words.Word{"CHANT", "ZZZZZ"},
			inputSolutions:   []words.Word{"SCARE", "SHARE", "SNARE", "STARE"},
			expected:         "CHANT",
			verboseError:     "Guessing %q would have reduced the shortlist to a single correct soluion, but Player guessed %q instead",
		},
		{
			name:             "Last turn",
			inputIsLastGuess: true,
			inputActionSpace: []words.Word{"CHANT", "SCARE"},
			inputSolutions:   []words.Word{"SCARE", "SHARE", "SNARE", "STARE"},
			expected:         "CHANT",
			verboseError:     "Because it's the last guess, guessing %q might have won the game, but Player guessed %q, which can't win",
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			p := NewPlayer(test.inputSolutions, test.inputActionSpace)
			got, _ := p.GetNextGuess(false)
			if !got.Equals(words.Word(test.expected)) {
				t.Errorf(test.verboseError, test.expected, got)
			}
		})
	}
}

func TestGuessEvaluation(t *testing.T) {

	table := []struct {
		name             string
		inputIsLastGuess bool
		inputActionSpace []words.Word
		inputSolutions   []words.Word
		expected         float64
		verboseError     string
	}{
		{
			name:             "Good guess",
			inputIsLastGuess: false,
			inputActionSpace: []words.Word{"CHANT"},
			inputSolutions:   []words.Word{"SCARE", "SHARE", "SNARE", "SPARE", "STARE"},
			expected:         0.2,
			verboseError:     "It reduces the shortlist to %.2f times its previous length with a given guess, but it seems to think the carry-over ratio is %.2f",
		},
		{
			name:             "Bad guess",
			inputIsLastGuess: false,
			inputActionSpace: []words.Word{"XXXXX"},
			inputSolutions:   []words.Word{"SCARE", "SHARE", "SNARE", "STARE"},
			expected:         1.0,
			verboseError:     "If you force it to choose a really unhelpful guess, shortlist carry-over ratio is %.2f, but it thinks it is %.2f",
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			p := NewPlayer(test.inputSolutions, test.inputActionSpace)
			_, evaluation := p.GetNextGuess(false)
			got := evaluation.GetWorstCaseShortlistCarryOverRatio()
			if got != test.expected {
				t.Errorf(test.verboseError, test.expected, got)
			}
		})
	}
}
