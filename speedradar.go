package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/net/context"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}

	signal.Ignore(syscall.SIGHUP)

	wg := sync.WaitGroup{}
	result := exportReadWrite{}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "sync", &wg)

	wg.Add(2)

	netToTerm := readAndWrite(ctx, conn, os.Stdout)
	termToNet := readAndWrite(ctx, os.Stdin, conn)

	select {
	case result = <-netToTerm:
	case result = <-termToNet:
	}

	if result.err != nil && result.err == io.EOF {
		result.err = nil
	}

	conn.Close()
	cancel()
	wg.Wait()
	fmt.Println(result.err)
}

type exportReadWrite struct {
	written uint64
	err     error
}

func readAndWrite(ctx context.Context, r io.Reader, w io.Writer) <-chan exportReadWrite {
	buff := make([]byte, 1024)
	c := make(chan exportReadWrite, 1)

	go func() {
		defer ctx.Value("sync").(*sync.WaitGroup).Done()

		export := exportReadWrite{}
		for {
			select {
			case <-ctx.Done():
				c <- export
				return
			default:
				nr, err := r.Read(buff)
				if err != nil {
					export.err = err
					c <- export
					return
				}
				if nr > 0 {
					wr, err := w.Write(buff[:nr])
					if err != nil {
						export.err = err
						c <- export
						return
					}
					if wr > 0 {
						export.written += uint64(wr)
					}
				}
			}
		}
	}()

	return c
}
