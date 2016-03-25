all: speedradar


speedradar: speedradar.go
	go build -o $@ $<
