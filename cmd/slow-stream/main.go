package main

import (
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/moul/slow-stream"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Manfred Touron"
	app.Email = "https://github.com/moul/slow-stream"
	// app.Version
	app.Usage = "Slow Stream"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode",
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}

	app.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			logrus.Debugf("pipe mode")

			pipe := slowstream.SlowStream(slowstream.SlowStreamOpts{
				Reader:       os.Stdin,
				Writer:       os.Stdout,
				MaxWriteSize: 1,
				WriteSleep:   100 * time.Millisecond,
			})
			var ret error
			select {
			case ret = <-pipe:
			}
			if ret != nil {
				logrus.Error(ret)
			}

		} else {
			logrus.Debugf("exec mode: %v", c.Args())
			wg := sync.WaitGroup{}

			wg.Add(2)

			spawn := exec.Command(c.Args()[0], c.Args()[1:]...)

			psOut, _ := spawn.StdoutPipe()
			psIn, _ := spawn.StdinPipe()
			defer psOut.Close()
			defer psIn.Close()

			psToTerm := slowstream.SlowStream(slowstream.SlowStreamOpts{
				Reader:       psOut,
				Writer:       os.Stdout,
				MaxWriteSize: 1,
				WriteSleep:   34 * time.Millisecond,
			})
			termToPs := slowstream.SlowStream(slowstream.SlowStreamOpts{
				Reader:       os.Stdin,
				Writer:       psIn,
				MaxWriteSize: 1,
				WriteSleep:   34 * time.Millisecond,
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
