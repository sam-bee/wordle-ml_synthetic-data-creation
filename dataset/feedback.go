package dataset

import (
	"fmt"

	"github.com/sam-bee/wordle-ml_game-engine/game"
	"github.com/sam-bee/wordle-ml_game-engine/words"
)

const feedbackPatternCount = 243

func feedbackCode(solution words.Word, guess words.Word) (uint8, error) {
	feedback := game.GetFeedback(solution, guess)
	return encodeFeedbackString(feedback.String())
}

func encodeFeedbackString(feedback string) (uint8, error) {
	if len(feedback) != MaxDepth {
		return 0, fmt.Errorf("feedback must be %d characters long, got %q", MaxDepth, feedback)
	}

	var code uint8
	var multiplier uint8 = 1
	for _, char := range feedback {
		value, err := feedbackCharValue(char)
		if err != nil {
			return 0, err
		}
		code += value * multiplier
		multiplier *= 3
	}
	return code, nil
}

func feedbackCharValue(char rune) (uint8, error) {
	switch char {
	case '-':
		return FeedbackGrey, nil
	case 'Y':
		return FeedbackYellow, nil
	case 'G':
		return FeedbackGreen, nil
	default:
		return 0, fmt.Errorf("unknown feedback character %q", char)
	}
}

func decodeFeedbackCode(code uint8) [MaxDepth]uint8 {
	var values [MaxDepth]uint8
	for i := range values {
		values[i] = code % 3
		code /= 3
	}
	return values
}
