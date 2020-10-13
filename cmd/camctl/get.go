package main

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
)

func getCommandFactory(ctx context.Context, ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return getCommand{
			ui:  ui,
			ctx: ctx,
		}, nil
	}
}

type getCommand struct {
	ui  cli.Ui
	ctx context.Context
}

func (g getCommand) Help() string {
	return `Usage: camctl get

Check whether the server thinks the camera is currently in use.
`
}

func (g getCommand) Synopsis() string {
	return "Check if the server thinks the camera is in use"
}

func (g getCommand) Run(args []string) int {
	status, err := getStatus(g.ctx)
	if err != nil {
		g.ui.Error(err.Error())
		return 1
	}
	if status == nil {
		g.ui.Output("This device hasn't checked in with the server yet.")
		return 0
	}
	statStr := "off"
	if status.CameraOn {
		statStr = "on"
	}
	g.ui.Output(fmt.Sprintf("As of %s, the server thinks this device is %s.", status.LastSync.Format("2006-01-02 15:04:05"), statStr))
	return 0
}
