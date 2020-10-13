package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"carvers.dev/camera-sign/device"
	"github.com/mitchellh/cli"
)

func checkCommandFactory(ctx context.Context, ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return checkCommand{
			ui:  ui,
			ctx: ctx,
		}, nil
	}
}

type checkCommand struct {
	ui  cli.Ui
	ctx context.Context
}

func (c checkCommand) Help() string {
	helpText := `Usage: camctl check [opts]`
	if runtime.GOOS == "windows" {
		helpText += ` [devices]`
	}
	helpText += `

Checks whether a camera is in use according to the options`
	if runtime.GOOS == "windows" {
		helpText += ` and devices`
	}
	helpText += `
specified, and reports the results to the server.

When -dry-run is specified, the result will only be printed to the terminal,
it will not be reported to the server.`
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

func (c checkCommand) Synopsis() string {
	return "Check whether a camera is in use"
}

func (c checkCommand) Run(args []string) int {
	// Windows can't list devices, so it can't check them all
	// so we need to tell it which specific devices to check.
	// It also has to check the handles for every single process
	// because it has no mapping of processes by handler. This
	// takes many seconds, and that's silly. We can speed it up
	// by giving it a list of process substrings to check against.
	var processes, devicePaths []string
	var dryRun bool
	var processesString string
	var err error

	f := flag.NewFlagSet("check", flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	// Set the default Usage to empty
	f.Usage = func() {}

	f.BoolVar(&dryRun, "dry-run", false, "should the result not be reported to the server")
	if runtime.GOOS == "windows" {
		f.StringVar(&processesString, "processes", "", "a comma-separated list of process substrings to limit your search to.")
	}

	f.Parse(args)

	var numArgs int
	if runtime.GOOS == "windows" {
		numArgs = 1
	}
	if len(f.Args()) != numArgs {
		c.ui.Error(fmt.Sprintf("Incorrect number of arguments. check command expects %d args, got %d.", numArgs, len(f.Args())))
		return 1
	}

	processes = strings.Split(processesString, ",")
	for pos, p := range processes {
		processes[pos] = strings.TrimSpace(p)
	}
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
			c.ui.Error("Error listing devices: " + err.Error())
			return 1
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
		inUse, err := device.InUse(c.ctx, dev, processes...)
		if err != nil {
			c.ui.Error("Error checking if device in use: " + err.Error())
			return 1
		}
		inUseStr := ""
		if !inUse {
			inUseStr = " not"
		}
		if dev != "" {
			c.ui.Info(fmt.Sprintf("Device %s is%s in use", dev, inUseStr))
		} else {
			c.ui.Info(fmt.Sprintf("Camera is%s in use", inUseStr))
		}
		if inUse {
			webcamOn = true
			break
		}
	}
	err = update(c.ctx, webcamOn)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	return 0
}
