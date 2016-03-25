package slowstream

import (
	"io"
	"time"
)

type SlowStreamOpts struct {
	Reader       io.Reader
	Writer       io.Writer
	MaxWriteSize int
	WriteSleep   time.Duration
}

func SlowStream(opts SlowStreamOpts) <-chan error {
	buff := make([]byte, 1024)
	c := make(chan error, 1)

	go func() {
		for {
			select {
			default:
				nr, err := opts.Reader.Read(buff)
				if err != nil {
					c <- err
					return
				}
				if nr > 0 {
					var end int
					for start := 0; start < nr; start = end {
						end = start + opts.MaxWriteSize
						if end > nr {
							end = nr
						}
						_, err := opts.Writer.Write(buff[start:end])
						if err != nil {
							c <- err
							return
						}
						if end == nr {
							break
						}
						time.Sleep(opts.WriteSleep)
					}
				}
			}
		}
	}()

	return c
}
