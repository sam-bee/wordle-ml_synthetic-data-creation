package dataset

import "testing"

func TestSplitSolutionsAddsMiniSubsetOfTrain(t *testing.T) {
	splits := splitSolutions(2309, 20260606)

	byID := make(map[SplitID]Split, len(splits))
	for _, split := range splits {
		byID[split.ID] = split
	}

	train := byID[SplitTrain]
	mini := byID[SplitMini]

	if len(mini.SolutionIDs) != miniSolutionCount {
		t.Fatalf("mini solution count = %d, want %d", len(mini.SolutionIDs), miniSolutionCount)
	}

	trainIDs := make(map[uint16]bool, len(train.SolutionIDs))
	for _, solutionID := range train.SolutionIDs {
		trainIDs[solutionID] = true
	}

	for _, solutionID := range mini.SolutionIDs {
		if !trainIDs[solutionID] {
			t.Fatalf("mini solution id %d is not in train", solutionID)
		}
	}
}
