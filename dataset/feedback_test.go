package dataset

import "testing"

func TestEncodeAndDecodeFeedbackString(t *testing.T) {
	code, err := encodeFeedbackString("-YGG-")
	if err != nil {
		t.Fatalf("encode feedback: %v", err)
	}

	got := decodeFeedbackCode(code)
	want := [MaxDepth]uint8{FeedbackGrey, FeedbackYellow, FeedbackGreen, FeedbackGreen, FeedbackGrey}
	if got != want {
		t.Fatalf("decoded feedback = %v, want %v", got, want)
	}
}

func TestEncodeFeedbackRejectsUnknownCharacters(t *testing.T) {
	if _, err := encodeFeedbackString("-YBG-"); err == nil {
		t.Fatal("expected an error for unknown feedback character")
	}
}
