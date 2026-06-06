# Synthetic Data Creation

This repository contains a Go CLI for creating synthetic Wordle imitation
learning data.

The generator uses the Wordle ML game engine and wordlists modules to create
incomplete game states, ask the teacher policy to rank legal next guesses, and
write training, validation, and test files under `data/`.

Generate the dataset with:

```text
go run . generate
```

Convert a binary dataset file to human-readable JSON with:

```text
go run . human-readable data/wordle-test.bin
```

Readable JSON output is written under `data/human-readable/`, which is ignored
by git.
