package dataset

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteBinaryFileUsesExpectedHeaderAndRecordSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wordle-train.bin")
	record, err := NewRecord(
		PaddingSolutionID,
		nil,
		[]uint16{0, 1},
		[]Label{{GuessID: 0, ReductionRatio: 0.5, WorstCaseSize: 1}},
		"test",
	)
	if err != nil {
		t.Fatalf("new record: %v", err)
	}

	config := Config{TopK: FixedTopK, MaxDepth: MaxDepth}
	if err := WriteBinaryFile(path, SplitTrain, []Record{record}, 3, 2, config); err != nil {
		t.Fatalf("write binary file: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read binary file: %v", err)
	}

	if len(data) != HeaderSizeBytes+RecordSizeBytes {
		t.Fatalf("file size = %d, want %d", len(data), HeaderSizeBytes+RecordSizeBytes)
	}
	if string(data[:4]) != FormatMagic {
		t.Fatalf("magic = %q, want %q", string(data[:4]), FormatMagic)
	}

	fields := data[4:HeaderSizeBytes]
	assertUint32Field(t, fields, 0, FormatVersion, "version")
	assertUint32Field(t, fields, 4, 1, "record_count")
	assertUint32Field(t, fields, 8, FixedTopK, "top_k")
	assertUint32Field(t, fields, 12, MaxDepth, "max_turns")
	assertUint32Field(t, fields, 16, 3, "guess_vocab_size")
	assertUint32Field(t, fields, 20, 2, "solution_count")
	assertUint32Field(t, fields, 24, uint32(SplitTrain), "split_id")
}

func assertUint32Field(t *testing.T, fields []byte, offset int, want uint32, name string) {
	t.Helper()
	got := binary.LittleEndian.Uint32(fields[offset : offset+4])
	if got != want {
		t.Fatalf("%s = %d, want %d", name, got, want)
	}
}
