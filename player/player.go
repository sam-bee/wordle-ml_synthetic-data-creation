package player

import (
	"runtime"
	"strings"

	"github.com/sam-bee/wordle-ml_game-engine/game"
	"github.com/sam-bee/wordle-ml_game-engine/words"
)

type Player struct {
	solutionShortlist []words.Word
	validGuesses      []words.Word
}

func NewPlayer(validSolutions []words.Word, validGuesses []words.Word) Player {
	return Player{solutionShortlist: validSolutions, validGuesses: validGuesses}
}

func (player *Player) GetNextGuess(lastTurn bool) (words.Word, GuessEvaluation) {

	if len(player.solutionShortlist) == 1 || lastTurn {
		return player.solutionShortlist[0], GuessEvaluation{Guess: player.solutionShortlist[0]}
	}

	bestGuess := player.identifyBestPossibleGuess(player.validGuesses)

	return bestGuess.Guess, bestGuess
}

func (player *Player) evaluateGuess(guess words.Word) GuessEvaluation {

	evaluation := NewGuessEvaluation(guess, player.solutionShortlist)

	for _, possibleSolution := range player.solutionShortlist {
		feedback := game.GetFeedback(possibleSolution, guess)
		evaluation.AddPossibleOutcome(possibleSolution, feedback)
	}

	return evaluation
}

func (p *Player) TakeFeedbackFromGuess(word words.Word, feedback game.Feedback) {

	var shortlist []words.Word

	for _, possibleSolution := range p.solutionShortlist {
		possibleFeedback := game.GetFeedback(possibleSolution, word)
		if possibleFeedback.Equals(feedback) {
			shortlist = append(shortlist, possibleSolution)
		}
	}

	p.solutionShortlist = shortlist
}

func (p *Player) ShortlistLength() int {
	return len(p.solutionShortlist)
}

func (p *Player) GetPossibleSolutions() string {
	var w []string
	for _, wo := range p.solutionShortlist {
		w = append(w, string(wo))
	}
	return strings.Join(w, ", ")
}

func fanoutGuessEvaluation(potentialGuesses []words.Word) <-chan words.Word {
	fanoutChannel := make(chan words.Word)
	go func() {
		for _, g := range potentialGuesses {
			fanoutChannel <- g
		}
		close(fanoutChannel)
	}()
	return fanoutChannel
}

func (p *Player) evaluateGuesses(fanoutChannel <-chan words.Word) <-chan GuessEvaluation {
	faninChannel := make(chan GuessEvaluation)
	go func() {

		bestGuess := NewEmptyEvaluation()

		for word := range fanoutChannel {
			evaluation := p.evaluateGuess(word)
			if evaluation.isBetterThan(bestGuess) {
				bestGuess = evaluation
			}
		}

		faninChannel <- bestGuess
		close(faninChannel)
	}()
	return faninChannel
}

func (player *Player) identifyBestPossibleGuess(validGuesses []words.Word) GuessEvaluation {

	// To fan out the guesses to the workers, create a fan out channel
	fanoutChannel := fanoutGuessEvaluation(validGuesses)

	// To collate the results from the workers, create one fan in channel per worker
	noOfWorkers := max(runtime.NumCPU()-1, 1)
	fanInChannels := make([]<-chan GuessEvaluation, noOfWorkers)

	for i := 0; i < noOfWorkers; i++ {
		fanInChannels[i] = player.evaluateGuesses(fanoutChannel)
	}

	// To identify the best guess from any of the workers, loop through their channels
	best := NewEmptyEvaluation()
	for i := range fanInChannels {
		for guess := range fanInChannels[i] {
			if guess.isBetterThan(best) {
				best = guess
			}
		}
	}

	return best
}
