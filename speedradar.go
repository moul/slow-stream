package main

import (
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/context"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}

	signal.Ignore(syscall.SIGHUP)

	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "sync", &wg)

	wg.Add(2)

	netToTerm := readAndWrite(readWriteOpts{
		Context:      ctx,
		Reader:       conn,
		Writer:       os.Stdout,
		MaxWriteSize: 1,
		WriteSleep:   34 * time.Millisecond,
	})
	termToNet := readAndWrite(readWriteOpts{
		Context:      ctx,
		Reader:       os.Stdin,
		Writer:       conn,
		MaxWriteSize: 1,
		WriteSleep:   34 * time.Millisecond,
	})

	var ret error
	select {
	case ret = <-netToTerm:
	case ret = <-termToNet:
	}

	if ret == io.EOF {
		ret = nil
	}

	conn.Close()
	cancel()
	wg.Wait()
	if ret != nil {
		panic(ret)
	}
}

type readWriteOpts struct {
	Context      context.Context
	Reader       io.Reader
	Writer       io.Writer
	MaxWriteSize int
	WriteSleep   time.Duration
}

func readAndWrite(opts readWriteOpts) <-chan error {
	buff := make([]byte, 1024)
	c := make(chan error, 1)

	go func() {
		defer opts.Context.Value("sync").(*sync.WaitGroup).Done()

		for {
			select {
			case <-opts.Context.Done():
				c <- nil
				return
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
