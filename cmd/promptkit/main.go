package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/promptkit/promptkit/internal/appdir"
	"github.com/promptkit/promptkit/internal/control"
	"github.com/promptkit/promptkit/internal/daemon"
	"github.com/promptkit/promptkit/internal/list"
	"github.com/promptkit/promptkit/internal/tui"
	"github.com/promptkit/promptkit/internal/view"
	cli "github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
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
				Name:   "ui",
				Usage:  "launch UI and control server",
				Action: uiCmd,
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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func startDaemon(_ context.Context, cmd *cli.Command) error {
	addr := cmd.String("addr")
	backend := cmd.String("backend")
	return daemon.Run(addr, backend)
}

func listCmd(_ context.Context, cmd *cli.Command) error {
	dir, err := appdir.SessionsDir()
	if err != nil {
		return err
	}

	sessions, err := list.LoadSessions(dir)
	if err != nil {
		return err
	}

	pred, err := list.ParseFilter(cmd.String("filter"))
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

	if limit := cmd.Int("limit"); limit > 0 && limit < len(summaries) {
		summaries = summaries[:limit]
	}

	if cmd.String("output") == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summaries)
	}

	list.PrintTable(summaries)
	return nil
}

func viewCmd(_ context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		return cli.Exit("session id required", 1)
	}
	id := cmd.Args().First()
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

func uiCmd(_ context.Context, cmd *cli.Command) error {
	addr := "localhost:5140"
	srv, err := control.NewServer(addr)
	if err != nil {
		return err
	}

	if err := srv.Start(); err != nil {
		return err
	}

	fmt.Println("🚀 Starting PromptKit control server on http://" + addr)
	fmt.Println("✅ Control server ready")

	ui := tui.New(addr)
	p := tea.NewProgram(ui)
	return p.Start()
}
