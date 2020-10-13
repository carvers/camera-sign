package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"carvers.dev/camera-sign/device"
	"github.com/mitchellh/cli"
)

func watchCommandFactory(ctx context.Context, ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return watchCommand{
			ui:  ui,
			ctx: ctx,
		}, nil
	}
}

type watchCommand struct {
	ui  cli.Ui
	ctx context.Context
}

func (w watchCommand) Help() string {
	helpText := `Usage: camctl watch [opts]`
	if runtime.GOOS == "windows" {
		helpText += ` [devices]`
	}
	helpText += `

Continuously checks whether a camera is in use according to the options`
	if runtime.GOOS == "windows" {
		helpText += ` and
devices`
	} else {
		helpText += `
`
	}
	helpText += `specified, and reports the results to the server.

When -cycle-time is set to a duration, the webcam status will be checked
with that duration. By default, it is checked every minute.`
	if runtime.GOOS == "windows" {
		helpText += `

When -processes is specified, it accepts a comma-separated list of process
substrings to search for. Only those processes usage of the webcam will be
reported. This speeds up the command considerably.

devices should be a comma-separated list of Physical Device Object names,
retrieved from Device Manager.`
	}
	return helpText
}

func (w watchCommand) Synopsis() string {
	return "Continuously check whether a camera is in use"
}

func (w watchCommand) Run(args []string) int {
	// Windows can't list devices, so it can't check them all
	// so we need to tell it which specific devices to check.
	// It also has to check the handles for every single process
	// because it has no mapping of processes by handler. This
	// takes many seconds, and that's silly. We can speed it up
	// by giving it a list of process substrings to check against.
	var processes, devicePaths []string
	var processesString string
	var err error
	var cycleTime time.Duration

	f := flag.NewFlagSet("watch", flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	// Set the default Usage to empty
	f.Usage = func() {}

	f.DurationVar(&cycleTime, "check-every", time.Minute, "how often to check webcam status, as a duration.")

	if runtime.GOOS == "windows" {
		f.StringVar(&processesString, "processes", "", "a comma-separated list of process substrings to limit your search to.")
	}

	f.Parse(args)

	if runtime.GOOS == "windows" {
		processes = strings.Split(processesString, ",")
		for pos, p := range processes {
			processes[pos] = strings.TrimSpace(p)
		}
	}

	var numArgs int
	if runtime.GOOS == "windows" {
		numArgs = 1
	}
	if len(f.Args()) != numArgs {
		w.ui.Error(fmt.Sprintf("Incorrect number of arguments. watch command expects %d args, got %d.", numArgs, len(f.Args())))
		return 1
	}

	for {

		if runtime.GOOS == "windows" {
			if len(f.Args()) > 0 {
				devicePaths = strings.Split(f.Args()[0], ",")
				for pos, p := range devicePaths {
					devicePaths[pos] = strings.TrimSpace(p)
				}
			}
		} else if runtime.GOOS == "linux" {
			devicePaths, err = filepath.Glob("/dev/video*")
			if err != nil {
				w.ui.Error("Error listing devices: " + err.Error())
				time.Sleep(time.Second)
				continue
			}
		} else if runtime.GOOS == "darwin" {
			// darwin doesn't use files or handlers to check
			// we just call out to a binary that checks all our
			// devices for us. So we just set a single empty
			// path, because it's not going to be used anyways,
			// and setting only one makes sure we call the code
			// just the once.
			devicePaths = append(devicePaths, "")
		}

		var webcamOn bool
		for _, dev := range devicePaths {
			inUse, err := device.InUse(w.ctx, dev, processes...)
			if err != nil {
				w.ui.Error("Error checking if device in use: " + err.Error())
				time.Sleep(time.Second)
				continue
			}
			inUseStr := ""
			if !inUse {
				inUseStr = " not"
			}
			if dev != "" {
				w.ui.Info(fmt.Sprintf("Device %s is%s in use", dev, inUseStr))
			} else {
				w.ui.Info(fmt.Sprintf("Camera is%s in use", inUseStr))
			}
			if inUse {
				webcamOn = true
				break
			}
		}
		err = update(w.ctx, webcamOn)
		if err != nil {
			w.ui.Error(err.Error())
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(cycleTime)
	}
	return 0
}
