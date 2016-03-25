all: slow-stream


slow-stream: ./cmd/slow-stream/main.go ./slow-stream.go
	go build -o $@ ./cmd/slow-stream/main.go
