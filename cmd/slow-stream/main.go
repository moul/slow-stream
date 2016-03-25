package main

import (
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"

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

		if len(c.Args()) == 0 {
			logrus.Debugf("pipe mode (buf=%d, dur=%v)", buffSize, maxWriteInterval)

			pipe := slowstream.SlowStream(slowstream.SlowStreamOpts{
				Reader:           os.Stdin,
				Writer:           os.Stdout,
				BuffSize:         buffSize,
				MaxWriteInterval: maxWriteInterval,
			})
			var ret error
			select {
			case ret = <-pipe:
			}
			if ret != nil {
				logrus.Error(ret)
			}

		} else {

			logrus.Debugf("exec mode: %v (buf=%d, dur=%v)", c.Args(), buffSize, maxWriteInterval)
			wg := sync.WaitGroup{}

			wg.Add(2)

			spawn := exec.Command(c.Args()[0], c.Args()[1:]...)

			psOut, _ := spawn.StdoutPipe()
			psIn, _ := spawn.StdinPipe()
			defer psOut.Close()
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
			psToTerm := slowstream.SlowStream(opts)

			// term to ps
			termToPs := slowstream.SlowStream(slowstream.SlowStreamOpts{
				Reader:           os.Stdin,
				Writer:           psIn,
				BuffSize:         buffSize,
				MaxWriteInterval: maxWriteInterval,
			})
			spawn.Stderr = os.Stderr

			spawn.Start()

			var ret error
			select {
			case ret = <-psToTerm:
			case ret = <-termToPs:
			}

			if ret == io.EOF {
				ret = nil
			}

			spawn.Wait()

			wg.Wait()
			if ret != nil {
				panic(ret)
			}
		}
	}

	app.Run(os.Args)
}
