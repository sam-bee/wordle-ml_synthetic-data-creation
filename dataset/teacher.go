package dataset

import (
	"fmt"
)

type Label struct {
	GuessID        uint16
	ReductionRatio float32
	WorstCaseSize  uint16
}

type Teacher struct {
	vocab  *Vocabulary
	matrix *FeedbackMatrix
}

func NewTeacher(vocab *Vocabulary, matrix *FeedbackMatrix) *Teacher {
	return &Teacher{vocab: vocab, matrix: matrix}
}

func (t *Teacher) Rank(shortlist []uint16, topK int) ([]Label, error) {
	if len(shortlist) == 0 {
		return nil, fmt.Errorf("cannot rank guesses for an empty solution shortlist")
	}
	if topK < 1 || topK > FixedTopK {
		return nil, fmt.Errorf("topK must be between 1 and %d, got %d", FixedTopK, topK)
	}
	if len(t.vocab.Guesses) < topK {
		return nil, fmt.Errorf("guess vocabulary has %d words, but topK is %d", len(t.vocab.Guesses), topK)
	}

	shortlistMembership := make([]bool, len(t.vocab.Solutions))
	for _, solutionID := range shortlist {
		shortlistMembership[solutionID] = true
	}

	top := make([]rankedLabel, 0, topK)
	var counts [feedbackPatternCount]uint16
	touched := make([]uint8, 0, feedbackPatternCount)

	for guessID := range t.vocab.Guesses {
		worstCaseSize := uint16(0)
		row := t.matrix.GuessRow(uint16(guessID))

		for _, solutionID := range shortlist {
			code := row[solutionID]
			if counts[code] == 0 {
				touched = append(touched, code)
			}
			counts[code]++
			if counts[code] > worstCaseSize {
				worstCaseSize = counts[code]
			}
		}

		for _, code := range touched {
			counts[code] = 0
		}
		touched = touched[:0]

		label := rankedLabel{
			Label: Label{
				GuessID:        uint16(guessID),
				ReductionRatio: 1 - float32(worstCaseSize)/float32(len(shortlist)),
				WorstCaseSize:  worstCaseSize,
			},
			isPotentialSolution: t.isPotentialSolutionGuess(uint16(guessID), shortlistMembership),
		}
		insertTopLabel(&top, label, topK)
	}

	labels := make([]Label, len(top))
	for i, label := range top {
		labels[i] = label.Label
	}

	return labels, nil
}

func (t *Teacher) isPotentialSolutionGuess(guessID uint16, shortlistMembership []bool) bool {
	solutionID := t.vocab.GuessSolutionID(guessID)
	return solutionID >= 0 && shortlistMembership[solutionID]
}

type rankedLabel struct {
	Label
	isPotentialSolution bool
}

func insertTopLabel(top *[]rankedLabel, candidate rankedLabel, limit int) {
	items := *top
	if len(items) == 0 {
		*top = append(items, candidate)
		return
	}

	insertAt := len(items)
	for i, item := range items {
		if candidate.betterThan(item) {
			insertAt = i
			break
		}
	}

	if insertAt >= limit {
		return
	}

	items = append(items, rankedLabel{})
	copy(items[insertAt+1:], items[insertAt:])
	items[insertAt] = candidate
	if len(items) > limit {
		items = items[:limit]
	}
	*top = items
}

func (label rankedLabel) betterThan(other rankedLabel) bool {
	if label.WorstCaseSize != other.WorstCaseSize {
		return label.WorstCaseSize < other.WorstCaseSize
	}
	if label.isPotentialSolution != other.isPotentialSolution {
		return label.isPotentialSolution
	}
	return label.GuessID < other.GuessID
}
