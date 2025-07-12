# PromptKit

PromptKit is a developer toolkit for recording and replaying prompt sessions with LLM APIs.

This repository contains a minimal prototype with a single command:

- `promptkit` – a CLI built with `urfave/cli/v2` that can start the daemon and will eventually manage sessions.

This is only a starting point and does not implement the full feature set.

## Running the Project

To run the project, use the following commands:

1. Start the daemon using the CLI:

   ```bash
   go run cmd/promptkit/main.go start
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
