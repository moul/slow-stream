package slowstream

import (
	"io"
	"sync"
	"time"

	"golang.org/x/net/context"
)

var VERSION = "1.2.0"

type SlowStreamOpts struct {
	Reader           io.Reader
	Writer           io.Writer
	BuffSize         int
	MaxWriteInterval time.Duration
}

func SlowStream(ctx context.Context, opts SlowStreamOpts) <-chan error {
	c := make(chan error, 1)

	go func() {
		var running chan error
		var fetch <-chan time.Time
		buff := make([]byte, opts.BuffSize)

		running = nil
		defer ctx.Value("sync").(*sync.WaitGroup).Done()
		for {
			if running == nil {
				fetch = time.After(0 * time.Second)
			}
			select {
			case <-ctx.Done():
				return
			case errRunning := <-running:
				running = nil
				if errRunning != nil {
					c <- errRunning
					return
				}
			case <-fetch:
				running = make(chan error, 1)
				go func() {
					nr, err := opts.Reader.Read(buff)
					if err != nil {
						running <- err
						return
					}

					if nr == 0 {
						running <- nil
						return
					}
					wr, err := opts.Writer.Write(buff[:nr])
					if err != nil {
						running <- err
						return
					}
					if wr > 0 {
						time.Sleep(opts.MaxWriteInterval)
					}
					running <- nil
				}()
			}
		}
	}()

	return c
}
