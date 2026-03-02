# AGENTS.md - Development Guidelines for TocadorMusica

## Project Overview

Music player project (TocadorMusica = MusicPlayer in Portuguese). The codebase is empty, so these are foundational guidelines to follow when building this project.

---

## Build, Lint, and Test Commands

### General Commands 
- `go get` - Install dependencies
- `go build -race` - Compile Project
- `go run -race` - Run in development

---

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
- Avoid relative imports beyond 2 levels (use path aliases)

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

## Recommended Project Structure

```
framework/
main.go
```

---

## Common Frameworks Used
- **github.com/ebitengine/oto/v3** - Music player
---

## Environment Variables Required

Create a `.config` file with:
```
```
