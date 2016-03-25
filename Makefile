all: slow-stream


slow-stream: slow-stream.go
	go build -o $@ $<
