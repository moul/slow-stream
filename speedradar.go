package main

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Manfred Touron"
	app.Email = "https://github.com/moul/speedradar"
	// app.Version
	app.Usage = "Speed Radar"

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

			pipe := SpeedRadar(SpeedRadarOpts{
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
		}
	}

	app.Run(os.Args)
}

type SpeedRadarOpts struct {
	Reader       io.Reader
	Writer       io.Writer
	MaxWriteSize int
	WriteSleep   time.Duration
}

func SpeedRadar(opts SpeedRadarOpts) <-chan error {
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
