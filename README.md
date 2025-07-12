# PromptKit

PromptKit is a developer toolkit for recording and replaying prompt sessions with LLM APIs.

This repository contains a minimal prototype with two commands:

- `promptkitd` – a simple reverse proxy that forwards requests to an OpenAI compatible backend and logs the request/response pairs to a JSON Lines file.
- `promptctl` – a CLI placeholder using `urfave/cli/v2` that will eventually control the daemon and manage sessions.

This is only a starting point and does not implement the full feature set.

## Running the Project

To run the project, use the following commands:

1. Start the daemon:

   ```bash
   go run cmd/promptkitd/main.go
   ```

2. Use the CLI to manage the daemon:

   ```bash
   go run cmd/promptctl/main.go
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

# Build promptkitd

go build -o bin/promptkitd cmd/promptkitd/main.go

# Build promptctl

go build -o bin/promptctl cmd/promptctl/main.go

```
