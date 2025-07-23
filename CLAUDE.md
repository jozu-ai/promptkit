# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building
```bash
go build -o bin/promptkit cmd/promptkit/main.go
```

### Testing
```bash
go test ./...
```

### Running the Application
```bash
# Start the daemon
go run cmd/promptkit/main.go start

# Launch the TUI
go run cmd/promptkit/main.go ui

# List sessions
go run cmd/promptkit/main.go list

# View session details
go run cmd/promptkit/main.go view <session-id>
```

## Architecture Overview

PromptKit is a local-first developer toolkit for recording and replaying prompt sessions with LLM APIs. The system operates as a reverse proxy that transparently captures OpenAI-compatible API interactions.

### Core Components

- **cmd/promptkit/main.go**: CLI entry point using `urfave/cli/v3` with commands for daemon management, TUI, and session operations
- **internal/daemon/**: HTTP server that acts as a reverse proxy, recording all LLM API traffic to local JSONL files
- **internal/tui/**: Bubble Tea terminal UI for browsing and managing recorded sessions
- **internal/recorder/**: Handles writing session data to disk in JSONL format
- **internal/proxy/**: HTTP proxy logic for forwarding requests to LLM backends
- **pkg/session/**: Core data structures for prompt sessions and metadata
- **internal/control/**: Control plane server for TUI communication
- **internal/list/**: Session listing and filtering functionality
- **internal/view/**: Session detail viewing and JSON formatting

### Data Flow

1. Client makes OpenAI-compatible API request to PromptKit daemon (default :8080)
2. Daemon proxies request to configured backend (default: https://api.openai.com)
3. Response is captured and logged to `~/.promptkit/sessions/chat-YYYY-MM-DD.jsonl`
4. TUI can browse sessions via control server API on localhost:5140

### Session Format

Each session is stored as a single JSON line containing:
- `id`: Unique session identifier
- `origin`: Source of the session (manual, framework, modelkit)
- `request`: Complete API request payload
- `response`: Complete API response
- `metadata`: Timestamp, latency, tags, publication status, content hash

### Key Dependencies

- **Bubble Tea**: Terminal UI framework for the TUI component
- **Chi**: HTTP router for the daemon and control server
- **urfave/cli/v3**: Command-line interface framework
- **fsnotify**: File system watching for session updates

### Testing Approach

The codebase includes focused unit tests for core functionality:
- `internal/daemon/handler_test.go`: HTTP handler testing
- `internal/list/list_test.go`: Session filtering and listing
- `internal/view/view_test.go`: Session detail formatting

Run tests with standard Go tooling - no special test runners or frameworks beyond the standard library testing package.