# AGENTS.md - Development Guidelines for TocadorMusica

## Project Overview

Music player project (TocadorMusica = MusicPlayer in Portuguese). Play a music/playlist URL or search by title!
---

## Build, Lint, and Test Commands

### General Commands 
- `go get` - Install dependencies
- `go build -race` - Compile Project
- `go run -race` - Run in development

---

## Repository Structure

```
tocadormusica/
├── config/                   # Internal package
│   ├── config.go             # Implementation
│   └── config_test.go        # Test the implementation
├── models/                   # Internal process
└── services/                 # Access to external API
    ├── yt-dlp/               # Access to yt-dlp service
    ├── audio/                # Access to oto player
    └── logger.go             # Logger
```

## Code Style Guidelines

### General Principles
- Write clean, readable, and maintainable code
- Follow the single responsibility principle
- Keep functions small and focused (max 30-40 lines when possible)
- Use meaningful variable and function names

### Naming Conventions
- **Files**: kebab-case (e.g., `music-queue.go`, `sound.go`)
- **Classes**: PascalCase (e.g., `MusicQueue`, `SoundHandler`)
- **Functions/variables**: camelCase (e.g., `getMusic`, `currentTrack`)
- **Constants**: SCREAMING_SNAKE_CASE (e.g., `MAX_QUEUE_SIZE`)
- **Interfaces**: PascalCase with `I` prefix optional (e.g., `Track` or `ITrack`)

### Imports
- Order imports consistently:
  1. External packages (e.g., `github.com/ebitengine/oto/v3`)
  2. Internal modules (e.g., `tocadormusica/commands`, `tocadormusica/utils`)
  3. Command line (e.g. `yt-dlp`, `ffmpeg`)
- Use absolute imports when configured

### Error Handling
- Always use try-catch for async operations
- Create custom error classes for domain-specific errors
- Log errors with appropriate context
- Never silently swallow errors
- Use result types (Either, Option) for functions that can fail

### Configuration
- Use `.config` file for secrets
- Validate configuration at startup

### Testing
- Write tests for business logic
- Use descriptive test names: `describe('MusicQueue', () => { it('should add track to queue', ...) })`
- Mock external dependencies (file system, network)
- Aim for meaningful test coverage, not just quantity

### Git Conventions
- Use conventional commits: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`
- Keep commits atomic and focused
- Write meaningful commit messages

---

## Common Frameworks Used
- **github.com/ebitengine/oto/v3** - Music player
---

## Environment Variables Required

Create a `.config` file with:
volume=0.01
search_results=10
```
```

## Adapters

