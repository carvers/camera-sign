package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/cli"
)

func setCommandFactory(ctx context.Context, ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return setCommand{
			ui:  ui,
			ctx: ctx,
		}, nil
	}
}

type setCommand struct {
	ui  cli.Ui
	ctx context.Context
}

func (s setCommand) Help() string {
	return `Usage: camctl set [state]

Manually sets whether a camera is in use according to the options.

state must be parseable as a boolean, or "off" or "on".`
}

func (s setCommand) Synopsis() string {
	return "Manually set whether a camera is in use"
}

func (s setCommand) Run(args []string) int {
	numArgs := 1
	if len(args) != numArgs {
		s.ui.Error(fmt.Sprintf("Incorrect number of arguments. check command expects %d args, got %d.", numArgs, len(args)))
		return 1
	}

	rawStatus := args[0]
	switch strings.ToLower(rawStatus) {
	case "on":
		rawStatus = "true"
	case "off":
		rawStatus = "false"
	}

	status, err := strconv.ParseBool(rawStatus)
	if err != nil {
		s.ui.Error("Error parsing camera state: " + err.Error())
		return 1
	}
	err = update(s.ctx, status)
	if err != nil {
		s.ui.Error(err.Error())
		return 1
	}
	return 0
}
