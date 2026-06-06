package player

import (
	"github.com/sam-bee/wordle-ml_game-engine/game"
	"github.com/sam-bee/wordle-ml_game-engine/words"
)

type GuessEvaluation struct {
	Guess                   words.Word
	shortlistSize           int
	potentialFeedbackCounts map[string]int
	isPotentialSolution     bool
}

func NewGuessEvaluation(guess words.Word, currrentShortlist []words.Word) GuessEvaluation {

	isPotentialSolution := false

	for _, wordInCurrentShortlist := range currrentShortlist {
		if wordInCurrentShortlist.Equals(guess) {
			isPotentialSolution = true
		}
	}

	return GuessEvaluation{
		Guess:                   guess,
		shortlistSize:           len(currrentShortlist),
		potentialFeedbackCounts: make(map[string]int),
		isPotentialSolution:     isPotentialSolution,
	}
}

func NewEmptyEvaluation() GuessEvaluation {
	return GuessEvaluation{}
}

func (ge *GuessEvaluation) AddPossibleOutcome(possibleSolution words.Word, feedback game.Feedback) {
	ge.potentialFeedbackCounts[feedback.String()] += 1
}

func (ge *GuessEvaluation) isBetterThan(another GuessEvaluation) bool {
	if ge.GetWorstCaseShortlistCarryOverRatio() < another.GetWorstCaseShortlistCarryOverRatio() {
		return true
	}
	if ge.GetWorstCaseShortlistCarryOverRatio() > another.GetWorstCaseShortlistCarryOverRatio() {
		return false
	}
	if ge.isPotentialSolution && !another.isPotentialSolution {
		return true
	}
	return false
}

func (ge *GuessEvaluation) GetWorstCaseFeedback() string {
	worstCaseFeedbackCount := 0
	var worstCaseFeedback string

	for feedback, feedbackCount := range ge.potentialFeedbackCounts {

		if feedbackCount > worstCaseFeedbackCount {
			worstCaseFeedbackCount = feedbackCount
			worstCaseFeedback = feedback
		}
	}

	return worstCaseFeedback
}

func (ge *GuessEvaluation) GetWorstCaseShortlistCarryOverRatio() float64 {
	worstCase := ge.GetWorstCaseFeedback()
	if worstCase == "" || ge.shortlistSize == 0 {
		return 1
	}
	return float64(ge.potentialFeedbackCounts[worstCase]) / float64(ge.shortlistSize)
}
