# Dataset Generator Design

The CLI generates the files described in `docs/wordle-imitation-dataset-files.md`.
Its purpose is to produce dense binary training data for a model that imitates
the current Go Wordle teacher.

## Dependencies

Wordle rules and word handling come from `github.com/sam-bee/wordle-ml_game-engine`.
The wordlist data comes from `github.com/sam-bee/wordle-ml_wordlists`. The
generator records a SHA-256 hash of the wordlist contents in every metadata
sidecar so a produced dataset can be traced back to the exact vocabulary.

## Records

Each solution word produces five records at each depth from 1 to 5. The training
file also includes one global opening-state record with no hidden solution. That
record uses `65535` as the solution ID sentinel.

Unused previous-guess slots use guess ID `65535`. Unused feedback slots use
feedback value `255`. Real feedback values are `0` for grey, `1` for yellow, and
`2` for green.

## State Generation

For each solution, the generator first lets the teacher play from the opening
state and keeps any incomplete states from that trajectory. Remaining records are
filled with random valid histories for the same solution and depth.

Random histories use legal guess words, never include the hidden solution as a
previous guess, and avoid repeated guesses within the same history. Every
feedback value is computed against the hidden solution, so every record remains
internally consistent even when the random history is not a likely teacher game.

The generation process is stochastic. The seed used for a run is stored in the
metadata, and each solution derives its own random stream from that seed so the
worker count does not affect the generated histories.

## Teacher Labels

The teacher ranks every legal guess by worst-case shortlist reduction. The stored
score is the reduction ratio, so higher is better. The binary record also stores
the worst-case shortlist size for each labelled guess.

The teacher breaks equal worst-case sizes by preferring guesses that are still
possible solutions, then by lower guess ID. Scores should be compared within a
single state rather than across unrelated states.

## Concurrency

The generator precomputes the feedback result for every solution/guess pair once,
then uses a fixed worker pool to generate solution records in parallel. The
default worker count follows Go's current processor setting, which is sensible
for this 16-thread desktop without creating nested worker pools.
