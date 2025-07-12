package main

import (
	"log"
	"os"
	"os/exec"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "promptctl",
		Usage: "manage promptkit",
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "start daemon",
				Action: startDaemon,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func startDaemon(c *cli.Context) error {
	cmd := exec.Command("./bin/promptkitd")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Starting promptkitd daemon...")
	if err := cmd.Start(); err != nil {
		return err
	}

	log.Printf("promptkitd daemon started with PID %d", cmd.Process.Pid)
	return nil
}
