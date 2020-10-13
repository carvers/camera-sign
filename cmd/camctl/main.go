package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/mitchellh/cli"
	"yall.in"
	"yall.in/colour"
)

func main() {
	// TODO: cancel context on interrupt
	ctx := context.Background()
	severity := os.Getenv("LOG_LEVEL")
	if severity == "" {
		severity = string(yall.Info)
	}
	switch yall.Severity(severity) {
	case yall.Default, yall.Debug, yall.Info, yall.Warning,
		yall.Error:
	default:
		fmt.Println("Unknown severity level %q; must be %q, %q, %q, or %q.",
			yall.Debug, yall.Info, yall.Warning, yall.Error)
	}
	logger := yall.New(colour.New(os.Stdout, yall.Severity(severity)))
	ctx = yall.InContext(ctx, logger)

	c := cli.NewCLI("cameractl", "0.1.0")
	c.Args = os.Args[1:]

	ui := &cli.ColoredUi{
		InfoColor:  cli.UiColorCyan,
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}

	c.Commands = map[string]cli.CommandFactory{
		"check": checkCommandFactory(ctx, ui),
		"watch": watchCommandFactory(ctx, ui),
		"set":   setCommandFactory(ctx, ui),
		"get":   getCommandFactory(ctx, ui),
	}

	exitStatus, err := c.Run()
	if err != nil {
		logger.WithError(err).Error("error running cameractl")
	}

	os.Exit(exitStatus)
}

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		if ifa.Flags&net.FlagUp == 0 || ifa.HardwareAddr == nil || ifa.Flags&net.FlagLoopback != 0 || ifa.Name == "docker0" {
			continue
		}
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}
