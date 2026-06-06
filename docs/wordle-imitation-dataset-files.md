# Wordle Imitation Learning Dataset Files

This document describes the proposed training, validation, and testing dataset files for the Wordle imitation learning project.

The purpose of these files is to train a neural policy model to imitate an existing Go Wordle player. The Go player acts as the teacher: given an incomplete Wordle state, it ranks legal next guesses according to worst-case shortlist reduction.

## File Set

The dataset should be split into three binary files:

```text
wordle-train.bin
wordle-validation.bin
wordle-test.bin
```

Each binary file should have a matching JSON metadata file with the same basename:

```text
wordle-train.json
wordle-validation.json
wordle-test.json
```

The `.bin` files are the fast path used by the trainer. The `.json` files are for inspection, reproducibility, and debugging. The trainer should not need to parse JSON in the hot path.

## Dataset Split

Each binary file should contain the same record structure. The difference between them is which solution words they contain.

The split should be by hidden solution word, not by individual training state.

For example, if the solution word `crane` is assigned to the test set, then no training or validation record should contain states generated from games whose hidden solution is `crane`.

Recommended split:

```text
training:    80% of valid solution words
validation:  10% of valid solution words
test:        10% of valid solution words
```

This gives a cleaner measure of whether the model generalises to unseen Wordle solutions.

## Opening State

The empty Wordle grid should not be repeated once per solution.

There should be a single opening-state record:

```text
no previous guesses
teacher top-16 opening guesses
teacher scores for those guesses
```

This record may be included in the training file only, or handled separately as a fixed opening policy case. It should not be duplicated across all solution shards.

## Per-Solution Composition

For each hidden solution word assigned to a split, generate 25 incomplete game states:

```text
depth 1: 5 states
depth 2: 5 states
depth 3: 5 states
depth 4: 5 states
depth 5: 5 states
```

Here, `depth` means the number of previous guesses already present in the Wordle state.

A depth-1 record asks:

```text
given one previous guess and its feedback, what should the next guess be?
```

A depth-5 record asks:

```text
given five previous guesses and their feedback, what should the sixth and final guess be?
```

## State Generation

For each solution, first allow the teacher to play a normal game. Any incomplete states from that teacher trajectory can be used to seed the relevant depth buckets.

For example, if the teacher solves a word in three guesses, that trajectory provides:

```text
depth 1 state
depth 2 state
```

The solved final guess is not itself an incomplete state requiring a next move.

Any remaining quota should be filled with generated valid unsolved histories:

```text
choose legal guesses
apply true Wordle feedback against the known hidden solution
stop at the target depth
reject histories where the solution has already been guessed
reject duplicate histories for the same solution/depth bucket
ask the teacher to rank the next legal guesses
```

The first implementation can fill missing states with random valid histories. Later versions may use a mixture of:

```text
teacher trajectory states
teacher-prefix plus random deviation
noisy teacher rollouts
random valid unsolved histories
```

The important invariant is that every record must describe a valid, internally consistent Wordle state for its hidden solution.

## Teacher Labels

For every generated state, the Go teacher should consider every legal guess and rank guesses by worst-case shortlist reduction.

Each record stores the top 16 teacher guesses:

```text
top_k = 16
```

For each top-k guess, store:

```text
guess word
worst-case shortlist reduction ratio
worst-case shortlist size after the guess
```

The reduction ratio should be in the range:

```text
0.0 to 1.0
```

Higher is better.

These scores are comparable within a single Wordle state, but should not be treated as globally comparable across unrelated games. The training loss should therefore normalise or compare scores within each individual record.

## Expected Dataset Size

Assuming approximately 2,500 valid solution words:

```text
2,500 solutions * 25 records per solution = 62,500 solution-specific records
```

With one global opening-state record:

```text
62,501 total records
```

With 16 labelled teacher actions per record:

```text
62,500 * 16 = 1,000,000 solution-specific labelled action scores
```

Including the opening state:

```text
62,501 * 16 = 1,000,016 labelled action scores
```

Under an 80/10/10 split, approximate record counts are:

```text
training:    ~50,000 solution-specific records
validation:  ~6,250 solution-specific records
test:        ~6,250 solution-specific records
```

The exact counts will depend on the final number of valid solution words.

## Binary File Format

The `.bin` files should be fixed-width, little-endian binary files.

JSON or CSV may be useful for inspection, but the main training files should be binary so they are easy for Go to write and efficient for the trainer to batch.

Each binary file should contain:

```text
header
records[]
```

## Header

Suggested header fields:

```text
magic bytes:       "WDIT"
version:           uint32
record_count:      uint32
top_k:             uint32
max_turns:         uint32
guess_vocab_size:  uint32
solution_count:    uint32
split_id:          uint32   // train, validation, or test
reserved:          bytes
```

The header should be explicit and versioned so future dataset formats can be detected safely.

## Record Structure

Current fixed-width record:

```text
solution_word[5]:               bytes
turn_depth:                     uint8

previous_guess_words[5][5]:     bytes
previous_feedback[5][5]:        uint8

shortlist_size_before:          uint16

top_k_guess_words[16][5]:       bytes
top_k_reduction_ratios[16]:     float32
top_k_worst_case_sizes[16]:     uint16
```

Word fields are fixed-width uppercase ASCII. Unused word fields and the global
opening-state solution word are padded with zero bytes.

Feedback values should use a small enum:

```text
0 = grey / absent
1 = yellow / present wrong position
2 = green / correct position
```

Unused previous-turn slots should be padded consistently.

For example, a depth-2 state uses:

```text
previous_guess_words[0]
previous_guess_words[1]
previous_feedback[0]
previous_feedback[1]
```

and pads slots 2, 3, and 4.

## JSON Metadata Sidecars

Every binary file should have a JSON sidecar file. The sidecar should describe the contents of that exact binary file.

For example:

```text
wordle-train.bin  -> wordle-train.json
wordle-test.bin   -> wordle-test.json
```

Suggested metadata:

```json
{
  "version": 1,
  "split": "train",
  "binary_file": "wordle-train.bin",
  "record_count": 50000,
  "top_k": 16,
  "max_turns": 5,
  "guess_vocab_size": 12972,
  "solution_count": 2000,
  "records_per_solution": 25,
  "records_per_depth": 5,
  "wordlist_hash": "...",
  "generator_commit": "...",
  "teacher_name": "worst_case_shortlist_reduction",
  "score_meaning": "Per-state worst-case shortlist reduction ratio. Higher is better. Not globally comparable across states."
}
```

The JSON should be treated as metadata only. The CUDA/GPU training path should consume dense batches assembled from the binary records.

## Validation and Test Usage

The training file is used for backpropagation.

The validation file is used while developing and training:

```text
monitor validation loss
detect overfitting
compare learning rates
compare model shapes
choose when to stop training
```

The test file should be kept untouched until the model design and training choices are final.

The final test should include both:

```text
static imitation metrics against the held-out test file
full Wordle game evaluation against held-out solution words
```

Useful static imitation metrics:

```text
teacher top-1 appears in model top-1
teacher top-1 appears in model top-5
teacher top-1 appears in model top-16
KL divergence or cross-entropy against the teacher top-k distribution
```

Useful game metrics:

```text
win rate
average guesses
failure count
distribution of guesses used
comparison against the original Go teacher
```
