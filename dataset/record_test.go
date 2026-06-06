package dataset

import "testing"

func TestNewRecordPadsUnusedHistoryAndLabels(t *testing.T) {
	record, err := NewRecord(
		7,
		[]Turn{{GuessID: 3, FeedbackCode: 5}},
		[]uint16{7, 8},
		[]Label{{GuessID: 4, ReductionRatio: 0.5, WorstCaseSize: 1}},
		"test",
	)
	if err != nil {
		t.Fatalf("new record: %v", err)
	}

	if record.TurnDepth != 1 {
		t.Fatalf("turn depth = %d, want 1", record.TurnDepth)
	}
	if record.PreviousGuessIDs[0] != 3 {
		t.Fatalf("first previous guess = %d, want 3", record.PreviousGuessIDs[0])
	}
	if record.PreviousGuessIDs[1] != PaddingGuessID {
		t.Fatalf("unused previous guess = %d, want padding", record.PreviousGuessIDs[1])
	}
	if record.PreviousFeedback[1][0] != PaddingFeedbackValue {
		t.Fatalf("unused feedback = %d, want padding", record.PreviousFeedback[1][0])
	}
	if record.TopKGuessIDs[0] != 4 {
		t.Fatalf("first label guess = %d, want 4", record.TopKGuessIDs[0])
	}
	if record.TopKGuessIDs[1] != PaddingGuessID {
		t.Fatalf("unused label guess = %d, want padding", record.TopKGuessIDs[1])
	}
}
