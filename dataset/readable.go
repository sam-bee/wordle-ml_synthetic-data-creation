package dataset

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const DefaultHumanReadableOutputDir = "data/human-readable"

type BinaryHeader struct {
	Magic          string  `json:"magic"`
	Version        uint32  `json:"version"`
	RecordCount    uint32  `json:"record_count"`
	TopK           uint32  `json:"top_k"`
	MaxTurns       uint32  `json:"max_turns"`
	GuessVocabSize uint32  `json:"guess_vocab_size"`
	SolutionCount  uint32  `json:"solution_count"`
	SplitID        SplitID `json:"split_id"`
	Split          string  `json:"split"`
}

type HumanReadableDataset struct {
	Header  BinaryHeader          `json:"header"`
	Records []HumanReadableRecord `json:"records"`
}

type HumanReadableRecord struct {
	Index                int                  `json:"index"`
	SolutionID           uint16               `json:"solution_id"`
	OpeningState         bool                 `json:"opening_state"`
	TurnDepth            uint8                `json:"turn_depth"`
	PreviousTurns        []HumanReadableTurn  `json:"previous_turns"`
	ShortlistSizeBefore  uint16               `json:"shortlist_size_before"`
	TeacherRankedGuesses []HumanReadableLabel `json:"teacher_ranked_guesses"`
}

type HumanReadableTurn struct {
	Turn           int    `json:"turn"`
	GuessID        uint16 `json:"guess_id"`
	Feedback       []int  `json:"feedback"`
	FeedbackString string `json:"feedback_string"`
}

type HumanReadableLabel struct {
	Rank           int     `json:"rank"`
	GuessID        uint16  `json:"guess_id"`
	ReductionRatio float32 `json:"reduction_ratio"`
	WorstCaseSize  uint16  `json:"worst_case_size"`
}

func WriteHumanReadableFile(binaryPath string, outputDir string) (string, error) {
	if outputDir == "" {
		outputDir = DefaultHumanReadableOutputDir
	}

	dataset, err := ReadHumanReadableDataset(binaryPath)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create human-readable output directory %q: %w", outputDir, err)
	}

	outputPath := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(binaryPath), filepath.Ext(binaryPath))+".json")
	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create human-readable file %q: %w", outputPath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dataset); err != nil {
		return "", fmt.Errorf("write human-readable file %q: %w", outputPath, err)
	}

	return outputPath, nil
}

func ReadHumanReadableDataset(path string) (HumanReadableDataset, error) {
	file, err := os.Open(path)
	if err != nil {
		return HumanReadableDataset{}, fmt.Errorf("open binary dataset %q: %w", path, err)
	}
	defer file.Close()

	header, err := readBinaryHeader(file)
	if err != nil {
		return HumanReadableDataset{}, fmt.Errorf("read header from %q: %w", path, err)
	}
	if err := validateHeader(header); err != nil {
		return HumanReadableDataset{}, fmt.Errorf("validate header from %q: %w", path, err)
	}

	records := make([]HumanReadableRecord, 0, header.RecordCount)
	for i := uint32(0); i < header.RecordCount; i++ {
		record, err := readHumanReadableRecord(file, int(i), header)
		if err != nil {
			return HumanReadableDataset{}, fmt.Errorf("read record %d from %q: %w", i, path, err)
		}
		records = append(records, record)
	}

	extra := make([]byte, 1)
	n, err := file.Read(extra)
	if err != nil && err != io.EOF {
		return HumanReadableDataset{}, fmt.Errorf("check for trailing bytes in %q: %w", path, err)
	}
	if n > 0 {
		return HumanReadableDataset{}, fmt.Errorf("binary dataset %q has trailing bytes after %d records", path, header.RecordCount)
	}

	return HumanReadableDataset{Header: header, Records: records}, nil
}

func readBinaryHeader(reader io.Reader) (BinaryHeader, error) {
	magic := make([]byte, len(FormatMagic))
	if _, err := io.ReadFull(reader, magic); err != nil {
		return BinaryHeader{}, err
	}

	var fields [7]uint32
	for i := range fields {
		if err := binary.Read(reader, binary.LittleEndian, &fields[i]); err != nil {
			return BinaryHeader{}, err
		}
	}

	reserved := make([]byte, HeaderSizeBytes-len(FormatMagic)-len(fields)*4)
	if _, err := io.ReadFull(reader, reserved); err != nil {
		return BinaryHeader{}, err
	}

	splitID := SplitID(fields[6])
	return BinaryHeader{
		Magic:          string(magic),
		Version:        fields[0],
		RecordCount:    fields[1],
		TopK:           fields[2],
		MaxTurns:       fields[3],
		GuessVocabSize: fields[4],
		SolutionCount:  fields[5],
		SplitID:        splitID,
		Split:          splitID.String(),
	}, nil
}

func validateHeader(header BinaryHeader) error {
	if header.Magic != FormatMagic {
		return fmt.Errorf("magic is %q, want %q", header.Magic, FormatMagic)
	}
	if header.Version != FormatVersion {
		return fmt.Errorf("version is %d, want %d", header.Version, FormatVersion)
	}
	if header.TopK != FixedTopK {
		return fmt.Errorf("top_k is %d, want %d", header.TopK, FixedTopK)
	}
	if header.MaxTurns != MaxDepth {
		return fmt.Errorf("max_turns is %d, want %d", header.MaxTurns, MaxDepth)
	}
	if header.Split == "unknown" {
		return fmt.Errorf("unknown split_id %d", header.SplitID)
	}
	return nil
}

func readHumanReadableRecord(reader io.Reader, index int, header BinaryHeader) (HumanReadableRecord, error) {
	record, err := readBinaryRecord(reader)
	if err != nil {
		return HumanReadableRecord{}, err
	}

	previousTurns := make([]HumanReadableTurn, 0, record.TurnDepth)
	for i := 0; i < int(record.TurnDepth); i++ {
		feedback := make([]int, 0, MaxDepth)
		for _, value := range record.PreviousFeedback[i] {
			feedback = append(feedback, int(value))
		}

		previousTurns = append(previousTurns, HumanReadableTurn{
			Turn:           i + 1,
			GuessID:        record.PreviousGuessIDs[i],
			Feedback:       feedback,
			FeedbackString: feedbackValuesString(feedback),
		})
	}

	labels := make([]HumanReadableLabel, 0, header.TopK)
	for i := 0; i < int(header.TopK); i++ {
		labels = append(labels, HumanReadableLabel{
			Rank:           i + 1,
			GuessID:        record.TopKGuessIDs[i],
			ReductionRatio: record.TopKReductionRatios[i],
			WorstCaseSize:  record.TopKWorstCaseSizes[i],
		})
	}

	return HumanReadableRecord{
		Index:                index,
		SolutionID:           record.SolutionID,
		OpeningState:         record.SolutionID == PaddingSolutionID,
		TurnDepth:            record.TurnDepth,
		PreviousTurns:        previousTurns,
		ShortlistSizeBefore:  record.ShortlistSizeBefore,
		TeacherRankedGuesses: labels,
	}, nil
}

func readBinaryRecord(reader io.Reader) (Record, error) {
	var record Record
	if err := binary.Read(reader, binary.LittleEndian, &record.SolutionID); err != nil {
		return Record{}, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &record.TurnDepth); err != nil {
		return Record{}, err
	}
	for i := range record.PreviousGuessIDs {
		if err := binary.Read(reader, binary.LittleEndian, &record.PreviousGuessIDs[i]); err != nil {
			return Record{}, err
		}
	}
	for turn := range record.PreviousFeedback {
		for position := range record.PreviousFeedback[turn] {
			if err := binary.Read(reader, binary.LittleEndian, &record.PreviousFeedback[turn][position]); err != nil {
				return Record{}, err
			}
		}
	}
	if err := binary.Read(reader, binary.LittleEndian, &record.ShortlistSizeBefore); err != nil {
		return Record{}, err
	}
	for i := range record.TopKGuessIDs {
		if err := binary.Read(reader, binary.LittleEndian, &record.TopKGuessIDs[i]); err != nil {
			return Record{}, err
		}
	}
	for i := range record.TopKReductionRatios {
		if err := binary.Read(reader, binary.LittleEndian, &record.TopKReductionRatios[i]); err != nil {
			return Record{}, err
		}
	}
	for i := range record.TopKWorstCaseSizes {
		if err := binary.Read(reader, binary.LittleEndian, &record.TopKWorstCaseSizes[i]); err != nil {
			return Record{}, err
		}
	}
	return record, nil
}

func feedbackValuesString(values []int) string {
	var builder strings.Builder
	for _, value := range values {
		switch uint8(value) {
		case FeedbackGrey:
			builder.WriteByte('-')
		case FeedbackYellow:
			builder.WriteByte('Y')
		case FeedbackGreen:
			builder.WriteByte('G')
		default:
			builder.WriteByte('?')
		}
	}
	return builder.String()
}
