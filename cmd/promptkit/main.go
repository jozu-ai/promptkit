package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/promptkit/promptkit/internal/appdir"
	"github.com/promptkit/promptkit/internal/daemon"
	"github.com/promptkit/promptkit/internal/list"
	"github.com/promptkit/promptkit/internal/view"
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
			{
				Name:        "list",
				Usage:       "list recorded sessions",
				Description: `List prompt sessions stored locally. The --filter flag accepts expressions like 'request.model=gpt-4', 'origin=modelkit', 'metadata.tags~qa', 'metadata.published!=null', 'metadata.timestamp>2025-07-01T00:00:00Z' or 'metadata.latency_ms<1000'.`,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "filter", Usage: "query expression using = != ~ < >, e.g. 'metadata.tags~qa'"},
					&cli.IntFlag{Name: "limit", Usage: "max results"},
					&cli.StringFlag{Name: "output", Value: "table", Usage: "output format (table|json)"},
				},
				Action: listCmd,
			},
			{
				Name:      "view",
				Usage:     "view session details",
				ArgsUsage: "<session-id>",
				Action:    viewCmd,
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

func listCmd(c *cli.Context) error {
	dir, err := appdir.SessionsDir()
	if err != nil {
		return err
	}

	sessions, err := list.LoadSessions(dir)
	if err != nil {
		return err
	}

	pred, err := list.ParseFilter(c.String("filter"))
	if err != nil {
		return err
	}

	var summaries []list.Summary
	for _, s := range sessions {
		smap := list.ToMap(s)
		if pred(smap) {
			summaries = append(summaries, list.Summarize(s))
		}
	}

	if limit := c.Int("limit"); limit > 0 && limit < len(summaries) {
		summaries = summaries[:limit]
	}

	if c.String("output") == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summaries)
	}

	list.PrintTable(summaries)
	return nil
}

func viewCmd(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("session id required", 1)
	}
	id := c.Args().First()
	dir, err := appdir.SessionsDir()
	if err != nil {
		return err
	}
	sess, err := view.FindSession(dir, id)
	if err != nil {
		return err
	}
	if sess == nil {
		fmt.Fprintf(os.Stderr, "❌ session '%s' not found\n", id)
		return cli.Exit("", 1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(sess)
}
