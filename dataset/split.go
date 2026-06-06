package dataset

import "math/rand"

type Split struct {
	ID          SplitID
	SolutionIDs []uint16
}

func splitSolutions(solutionCount int, seed int64) []Split {
	rng := rand.New(rand.NewSource(seed))
	permutation := rng.Perm(solutionCount)

	trainingCount := solutionCount * 80 / 100
	validationCount := solutionCount * 10 / 100

	splits := []Split{
		{ID: SplitTrain, SolutionIDs: make([]uint16, 0, trainingCount)},
		{ID: SplitValidation, SolutionIDs: make([]uint16, 0, validationCount)},
		{ID: SplitTest, SolutionIDs: make([]uint16, 0, solutionCount-trainingCount-validationCount)},
	}

	for i, solutionID := range permutation {
		switch {
		case i < trainingCount:
			splits[0].SolutionIDs = append(splits[0].SolutionIDs, uint16(solutionID))
		case i < trainingCount+validationCount:
			splits[1].SolutionIDs = append(splits[1].SolutionIDs, uint16(solutionID))
		default:
			splits[2].SolutionIDs = append(splits[2].SolutionIDs, uint16(solutionID))
		}
	}

	return splits
}
