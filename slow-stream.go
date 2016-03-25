package slowstream

import (
	"io"
	"time"
)

var VERSION = "1.1.0"

type SlowStreamOpts struct {
	Reader           io.Reader
	Writer           io.Writer
	BuffSize         int
	MaxWriteInterval time.Duration
}

func SlowStream(opts SlowStreamOpts) <-chan error {
	buff := make([]byte, opts.BuffSize)
	c := make(chan error, 1)

	go func() {
		for {
			nr, err := opts.Reader.Read(buff)
			if err != nil {
				c <- err
				return
			}
			if nr == 0 {
				continue
			}

			wr, err := opts.Writer.Write(buff[:nr])
			if err != nil {
				c <- err
				return
			}
			if wr > 0 {
				time.Sleep(opts.MaxWriteInterval)
			}
		}
	}()

	return c
}
