package main

import (
	"context"
	"errors"
	"github.com/ThinkerQAQ/reverfolio/internal/echo"
)

func runMCPCommand(args []string) error {
	command := args[0]
	rest := args[1:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == "--config" {
			if i+1 >= len(rest) {
				return errors.New("--config requires a path")
			}
			if err := echo.SetMCPConfigPath(rest[i+1]); err != nil {
				return err
			}
			rest = append(rest[:i], rest[i+2:]...)
			i--
		}
	}
	if command == "tunnel-run" {
		if len(rest) != 0 {
			return errors.New("unexpected Tunnel runner argument")
		}
		return echo.RunTunnel(context.Background())
	}
	return runMCPManagement(rest)
}
