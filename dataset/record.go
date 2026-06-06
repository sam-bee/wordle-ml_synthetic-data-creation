package dataset

import "fmt"

type Turn struct {
	GuessID      uint16
	FeedbackCode uint8
}

type Record struct {
	SolutionID             uint16
	TurnDepth              uint8
	PreviousGuessIDs       [MaxDepth]uint16
	PreviousFeedback       [MaxDepth][MaxDepth]uint8
	ShortlistSizeBefore    uint16
	TopKGuessIDs           [FixedTopK]uint16
	TopKReductionRatios    [FixedTopK]float32
	TopKWorstCaseSizes     [FixedTopK]uint16
	Source                 string
	HistoryDeduplicationID string
}

func NewRecord(solutionID uint16, history []Turn, shortlist []uint16, labels []Label, source string) (Record, error) {
	if len(history) > MaxDepth {
		return Record{}, fmt.Errorf("history depth %d exceeds max depth %d", len(history), MaxDepth)
	}
	if len(shortlist) > int(^uint16(0)) {
		return Record{}, fmt.Errorf("shortlist size %d exceeds uint16 record field", len(shortlist))
	}
	if len(labels) > FixedTopK {
		return Record{}, fmt.Errorf("label count %d exceeds fixed topK %d", len(labels), FixedTopK)
	}

	record := Record{
		SolutionID:             solutionID,
		TurnDepth:              uint8(len(history)),
		ShortlistSizeBefore:    uint16(len(shortlist)),
		Source:                 source,
		HistoryDeduplicationID: historyKey(history),
	}

	for i := range record.PreviousGuessIDs {
		record.PreviousGuessIDs[i] = PaddingGuessID
	}
	for turn := range record.PreviousFeedback {
		for position := range record.PreviousFeedback[turn] {
			record.PreviousFeedback[turn][position] = PaddingFeedbackValue
		}
	}
	for i := range record.TopKGuessIDs {
		record.TopKGuessIDs[i] = PaddingGuessID
	}

	for i, turn := range history {
		record.PreviousGuessIDs[i] = turn.GuessID
		record.PreviousFeedback[i] = decodeFeedbackCode(turn.FeedbackCode)
	}

	for i, label := range labels {
		record.TopKGuessIDs[i] = label.GuessID
		record.TopKReductionRatios[i] = label.ReductionRatio
		record.TopKWorstCaseSizes[i] = label.WorstCaseSize
	}

	return record, nil
}

func historyKey(history []Turn) string {
	key := make([]byte, 0, len(history)*5)
	for _, turn := range history {
		key = append(key, byte(turn.GuessID), byte(turn.GuessID>>8), turn.FeedbackCode, ';')
	}
	return string(key)
}
