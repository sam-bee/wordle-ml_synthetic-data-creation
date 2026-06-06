package dataset

import "fmt"

type FeedbackMatrix struct {
	vocab        *Vocabulary
	byGuess      []uint8
	solutionSize int
	guessSize    int
}

func NewFeedbackMatrix(vocab *Vocabulary) (*FeedbackMatrix, error) {
	matrix := &FeedbackMatrix{
		vocab:        vocab,
		solutionSize: len(vocab.Solutions),
		guessSize:    len(vocab.Guesses),
		byGuess:      make([]uint8, len(vocab.Guesses)*len(vocab.Solutions)),
	}

	for guessID, guess := range vocab.Guesses {
		rowOffset := guessID * matrix.solutionSize
		for solutionID, solution := range vocab.Solutions {
			code, err := feedbackCode(solution, guess)
			if err != nil {
				return nil, fmt.Errorf("encode feedback for solution %q and guess %q: %w", solution, guess, err)
			}
			matrix.byGuess[rowOffset+solutionID] = code
		}
	}

	return matrix, nil
}

func (m *FeedbackMatrix) FeedbackCode(solutionID uint16, guessID uint16) uint8 {
	return m.byGuess[int(guessID)*m.solutionSize+int(solutionID)]
}

func (m *FeedbackMatrix) GuessRow(guessID uint16) []uint8 {
	offset := int(guessID) * m.solutionSize
	return m.byGuess[offset : offset+m.solutionSize]
}

func (m *FeedbackMatrix) Shortlist(history []Turn) []uint16 {
	shortlist := make([]uint16, 0, m.solutionSize)

	for solutionID := 0; solutionID < m.solutionSize; solutionID++ {
		if m.solutionMatchesHistory(uint16(solutionID), history) {
			shortlist = append(shortlist, uint16(solutionID))
		}
	}

	return shortlist
}

func (m *FeedbackMatrix) solutionMatchesHistory(solutionID uint16, history []Turn) bool {
	for _, turn := range history {
		if m.FeedbackCode(solutionID, turn.GuessID) != turn.FeedbackCode {
			return false
		}
	}
	return true
}
