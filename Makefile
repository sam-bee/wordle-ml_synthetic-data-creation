.PHONY: compile test data human-readable

compile:
	go build -o wordle .

test:
	go test ./...

data:
	go run . generate --output data --seed 20260606 --workers 16

human-readable:
	go run . human-readable data/wordle-train.bin
	go run . human-readable data/wordle-mini.bin
	go run . human-readable data/wordle-validation.bin
	go run . human-readable data/wordle-test.bin
