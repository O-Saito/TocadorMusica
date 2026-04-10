# AGENTS.md - Development Guidelines for TocadorMusica

## Project Overview

Music player project (TocadorMusica = MusicPlayer in Portuguese). Play a music/playlist of a Youtube URL or Youtube search (displaying a list and make the user select) or a file
---

## Perfil
This application should be able to handle multi-perfil. Being handled by a abstraction
Every Perfil should run in a different goroutine.
Perfis are isolated. They do not share state.
Each Perfil must contain:
- Input handler (one or more)
- Playback queue
- Audio output (pick one)
- Context for cancellation
- Dedicated logger instance
- Command handler

## Queue
- FIFO
- Auto play next
- Supports: clear, remove one(index)
- All queue operations must be safe for concurrent use
- Queue must not block input handlers

## Build, Lint, and Test Commands

### General Commands 
- `go get` - Install dependencies
- `go build -race` - Compile Project
- `go run -race` - Run in development

---

## Repository Structure

```
tocadormusica/
├── main.go
├── domain/
├── ports/
│   ├── ui/
│   └── audio/                # As output
├── adapters/
│   ├── ui/
│   └── audio/                # As output
├── config/                   # Internal package
│   ├── config.go             # Implementation
│   └── config_test.go        # Test the implementation
├── logger/                   # Internal package
│   ├── logger.go             # Implementation
│   └── logget_test.go        # Test the implementation
└── services/                 # Access to external API
    ├── yt-dlp/               # Access to yt-dlp service as package
    ├── audio/                # Access to oto player as package
    └── logger.go             # Logger
```

## Code Style Guidelines

### General Principles
- Write clean, readable, and maintainable code
- Follow the single responsibility principle
- Keep functions small and focused (max 30-40 lines when possible)
- Use meaningful variable and function names

### Naming Conventions (Go)
- **Files**: snake_case (e.g., `music_queue.go`, `sound.go`)
- **Structs**: PascalCase (e.g., `MusicQueue`, `SoundHandler`)
- **Exported functions/types**: PascalCase (e.g., `MusicQueue`, `PlayTrack`)
- **Unexported functions/types**: camelCase (e.g., `loadConfig`)
- **Constants**: SCREAMING_SNAKE_CASE (e.g., `MAX_QUEUE_SIZE`)
- **Interfaces**: PascalCase with `I` prefix optional (e.g., `Track` or `ITrack`)
- **Packages**: File/Folder (e.g., `tocadormusica/new_commands`); The import (e.g.,`tocadormusica/new_commands`)
- **Avoid stutter**: `audio.Player` not `audio.AudioPlayer`

## Concurrency Rules
- Each Perfil runs in its own goroutine
- Communication between goroutines must use channels
- Shared state must be protected (mutex or avoided)
- No global mutable state
- Always support graceful shutdown (context.Context)
- Audio shouldn't block input thread

## Dependency Rules
- domain must not import adapters
- services may import domain and ports
- adapters implement ports
- main wires dependencies
- No circular dependencies allowed

## Architecture: Hexagonal (Ports & Adapters)

Core domain must NOT depend on:
- oto
- CLI

### Layers
- domain/         → Business rules
- services/       → Use cases
- ports/          → Interfaces
- adapters/       → External integrations

### Imports
- Order imports consistently:
  1. External packages (e.g., `github.com/ebitengine/oto/v3`)
  2. Internal modules (e.g., `tocadormusica/commands`, `tocadormusica/utils`)
  4. Command line (e.g. `yt-dlp`, `ffmpeg`, `deno`)
- Use absolute imports when configured

### Error Handling 
- Return errors as last return value
- Wrap errors using fmt.Errorf("context: %w", err)
- Define sentinel errors when needed
- Never panic in business logic
- Panic only in unrecoverable startup errors

### Configuration
- Use `.config` file for secrets
- Validate configuration at startup

### Testing
- Use Go standard `testing` package
- Test files: *_test.go
- Test functions: TestFunctionName(t *testing.T)
- Use table-driven tests
- Mock external services via interfaces
- No network calls in unit tests
- No real file system usage in unit tests
- Integration tests must be separated

### Git Conventions
- Use conventional commits: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`
- Keep commits atomic and focused
- Write meaningful commit messages

---

## Common Frameworks Used
- **github.com/ebitengine/oto/v3** - Music player (for local)
---

## Environment Variables Required

Create a `.config` file with:
volume=0.01
search_results=10
music_folders=
```
```

## Logging
- The log should have a file for the current application, getting the initial date-time of startup
- Log format: [2006-01-02T15:04:05] [goroutines=<number>] [LEVEL] [perfil] [file:line] message
- Logger must be injected (no global logger)
- Logger must be interface-based
- Log level configurable via .config
- Log level empty is no log at all (into the console)

### Track
```
TrackType YouTube|File

type Track struct {
  URL         string
  Title       string
  Description string
  AudioURL    string
  Type        TrackType
}
```

## Youtube
- The youtube data should be fetched by yt-dlp
- The audio stream should be fetched by ffmpeg
- The search should only happen when the input is not a youtube URL and was not found on a directory 
- Audio url is in Formats->When resolution is "audio only" get URL
- The data of a video should be 
```
  URL: Video URL
  Title: Video Title
  Description: Video Description
  AudioURL: audio url that contains in Formats->When resolution is "audio only" get URL
```

## File
- Should search in directory list (on config)
- If is a video should get only the audio
- The data of a audio should be 
```
  URL: File path
  Title: Folders of relative recursive folder + File name
  Description: Empty
  AudioURL: File path
```

## Performance & Stability
- No goroutine leaks
- All channels must be closed properly
- Audio playback must not block main thread
- Startup time < 2s
- Configuration/dependencies path should be resolved as relative to the executable location

## External Processes
- Must enforce timeouts
- Must not leak zombie processes

## External packages should have auto installs
Should ask if the user want a auto install of ffmpeg, yt-dlp and deno
