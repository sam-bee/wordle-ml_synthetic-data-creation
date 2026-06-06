package dataset

import (
	"fmt"
	"strings"

	"github.com/sam-bee/wordle-ml_game-engine/words"
)

type WordBytes [WordLength]byte

type Turn struct {
	GuessID      uint16
	FeedbackCode uint8
}

type Record struct {
	solutionID             uint16
	SolutionWord           WordBytes
	TurnDepth              uint8
	PreviousGuessWords     [MaxDepth]WordBytes
	PreviousFeedback       [MaxDepth][MaxDepth]uint8
	ShortlistSizeBefore    uint16
	TopKGuessWords         [FixedTopK]WordBytes
	TopKReductionRatios    [FixedTopK]float32
	TopKWorstCaseSizes     [FixedTopK]uint16
	Source                 string
	HistoryDeduplicationID string
}

func NewRecord(vocab *Vocabulary, solutionID uint16, history []Turn, shortlist []uint16, labels []Label, source string) (Record, error) {
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
		solutionID:             solutionID,
		TurnDepth:              uint8(len(history)),
		ShortlistSizeBefore:    uint16(len(shortlist)),
		Source:                 source,
		HistoryDeduplicationID: historyKey(history),
	}
	if solutionID != PaddingSolutionID {
		record.SolutionWord = wordBytesFromWord(vocab.Solutions[solutionID])
	}

	for turn := range record.PreviousFeedback {
		for position := range record.PreviousFeedback[turn] {
			record.PreviousFeedback[turn][position] = PaddingFeedbackValue
		}
	}

	for i, turn := range history {
		record.PreviousGuessWords[i] = wordBytesFromWord(vocab.Guesses[turn.GuessID])
		record.PreviousFeedback[i] = decodeFeedbackCode(turn.FeedbackCode)
	}

	for i, label := range labels {
		record.TopKGuessWords[i] = wordBytesFromWord(vocab.Guesses[label.GuessID])
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

func wordBytesFromWord(word words.Word) WordBytes {
	var encoded WordBytes
	copy(encoded[:], string(word))
	return encoded
}

func wordBytesString(word WordBytes) string {
	return strings.TrimRight(string(word[:]), "\x00")
}
