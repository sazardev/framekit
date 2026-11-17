BIN := framekit

.PHONY: build test run clean tidy

build:
	go build -o $(BIN) ./cmd/framekit

test:
	go test ./...

tidy:
	go mod tidy

run:
	go run ./cmd/framekit render --spec examples/intro.json --out /tmp/framekit-intro.mp4

clean:
	rm -f $(BIN)
