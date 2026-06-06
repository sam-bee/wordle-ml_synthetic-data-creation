package dataset

const (
	FormatMagic   = "WDIT"
	FormatVersion = 2

	FixedTopK  = 16
	MaxDepth   = 5
	WordLength = 5

	FeedbackGrey   uint8 = 0
	FeedbackYellow uint8 = 1
	FeedbackGreen  uint8 = 2

	PaddingSolutionID    uint16 = 0xffff
	PaddingFeedbackValue uint8  = 0xff

	HeaderSizeBytes = 64
	RecordSizeBytes = 234
)

type SplitID uint32

const (
	SplitTrain SplitID = iota + 1
	SplitValidation
	SplitTest
)

func (s SplitID) String() string {
	switch s {
	case SplitTrain:
		return "train"
	case SplitValidation:
		return "validation"
	case SplitTest:
		return "test"
	default:
		return "unknown"
	}
}

func (s SplitID) FileStem() string {
	switch s {
	case SplitTrain:
		return "wordle-train"
	case SplitValidation:
		return "wordle-validation"
	case SplitTest:
		return "wordle-test"
	default:
		return "wordle-unknown"
	}
}
