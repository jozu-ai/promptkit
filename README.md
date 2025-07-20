# PromptKit

PromptKit is a developer toolkit for recording and replaying prompt sessions with LLM APIs.

This repository contains a minimal prototype with a CLI and a terminal UI.

 - `promptkit` – a CLI built with `urfave/cli/v3` that can start the daemon and manage sessions.
- `promptkit ui` – launches a Bubble Tea TUI for browsing recorded sessions.

## Running the Project

To run the project, use the following commands:

1. Start the daemon using the CLI:

   ```bash
   go run cmd/promptkit/main.go start
   ```

2. Launch the TUI:

   ```bash
   go run cmd/promptkit/main.go ui
   ```

## Testing the Project

To test the project, run:

```bash

go test ./...

```

This will execute all tests in the repository.

## Building the Project

To build the binaries for the project, use:

```bash

# Build promptkit

go build -o bin/promptkit cmd/promptkit/main.go

```
