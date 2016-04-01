package main

import (
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/moul/slow-stream"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Manfred Touron"
	app.Email = "https://github.com/moul/slow-stream"
	app.Version = slowstream.VERSION
	app.Usage = "Slow Stream"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode",
		},
		cli.BoolFlag{
			Name:  "raw, r",
			Usage: "Enable raw mode",
		},
		cli.BoolFlag{
			Name:  "stdout-passthrough",
			Usage: "Do not slow stdout",
		},
		cli.IntFlag{
			Name:  "max-write-interval, i",
			Usage: "Max write interval (in millisecond)",
			Value: 100,
		},
		cli.IntFlag{
			Name:  "buff-size, b",
			Usage: "Buffer size",
			Value: 1024,
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}

	app.Action = func(c *cli.Context) {
		if c.Bool("raw") {
			oldState, err := terminal.MakeRaw(0)
			if err != nil {
				logrus.Fatal(err)
			}

			defer terminal.Restore(0, oldState)
		}

		buffSize := c.Int("buff-size")
		maxWriteInterval := time.Duration(c.Int("max-write-interval")) * time.Millisecond

		waitGroup := sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, "sync", &waitGroup)

		if len(c.Args()) == 0 {
			logrus.Debugf("pipe mode (buf=%d, dur=%v)", buffSize, maxWriteInterval)

			waitGroup.Add(1)
			if ret := <-slowstream.SlowStream(ctx, slowstream.SlowStreamOpts{
				Reader:           os.Stdin,
				Writer:           os.Stdout,
				BuffSize:         buffSize,
				MaxWriteInterval: maxWriteInterval,
			}); ret != nil && ret != io.EOF {
				logrus.Error(ret)
			}
			cancel()
			waitGroup.Wait()
		} else {
			logrus.Debugf("exec mode: %v (buf=%d, dur=%v)", c.Args(), buffSize, maxWriteInterval)

			spawn := exec.Command(c.Args()[0], c.Args()[1:]...)

			psOut, err := spawn.StdoutPipe()
			if err != nil {
				logrus.Fatal(err)
			}
			defer psOut.Close()

			psIn, err := spawn.StdinPipe()
			if err != nil {
				logrus.Error(err)
				return // don't use Fatal to call defer functions
			}
			defer psIn.Close()

			// ps to term
			opts := slowstream.SlowStreamOpts{
				Reader:           psOut,
				Writer:           os.Stdout,
				BuffSize:         buffSize,
				MaxWriteInterval: maxWriteInterval,
			}
			if c.Bool("stdout-passthrough") {
				opts.BuffSize = 1024
				opts.MaxWriteInterval = 0
			}
			waitGroup.Add(1)
			psToTerm := slowstream.SlowStream(ctx, opts)

			// term to ps
			waitGroup.Add(1)
			termToPs := slowstream.SlowStream(ctx, slowstream.SlowStreamOpts{
				Reader:           os.Stdin,
				Writer:           psIn,
				BuffSize:         buffSize,
				MaxWriteInterval: maxWriteInterval,
			})
			spawn.Stderr = os.Stderr

			if err := spawn.Start(); err != nil {
				logrus.Error(err)
				return // don't use Fatal to call defer functions
			}

			var ret error

			select {
			case ret = <-psToTerm:
			case ret = <-termToPs:
			}

			if ret == io.EOF {
				ret = nil
			}

			if err := spawn.Wait(); err != nil {
				logrus.Error(err)
				return // don't use Fatal to call defer functions
			}
			cancel()
			waitGroup.Wait()

			if ret != nil {
				logrus.Error(ret)
				return // don't use Fatal to call defer functions
			}
		}
	}

	app.Run(os.Args)
}
