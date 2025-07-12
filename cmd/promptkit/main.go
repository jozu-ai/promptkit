package main

import (
	"log"
	"os"

	"github.com/promptkit/promptkit/internal/daemon"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "promptkit",
		Usage: "manage promptkit",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start daemon",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "addr", Value: ":8080", Usage: "listen address"},
					&cli.StringFlag{Name: "backend", Value: "https://api.openai.com", Usage: "backend base URL"},
				},
				Action: startDaemon,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func startDaemon(c *cli.Context) error {
	addr := c.String("addr")
	backend := c.String("backend")
	return daemon.Run(addr, backend)
}
