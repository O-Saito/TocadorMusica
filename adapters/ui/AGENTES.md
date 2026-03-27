# CLIAdapter - Development Guidelines

## Overview

CLIAdapter is the terminal-based UI implementation for the music player. It handles user input and display output in a CLI environment using the lipgloss library for styling.

## File Location

```
adapters/ui/cli.go
```

## Dependencies

- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/lipgloss/table` - Table layouts
- `tocadormusica/ports/ui` - Interface definitions
- `tocadormusica/commands` - Command listing

## Display Layout

### Main Box (103 chars wide)

```
┌──────────────────────────────────────────────────────────────────┐
│                        TOCADOR DE MUSICA                         │
├──────────────────────────────────────────────────────────────────┤
│ "Track Title"                                                    │
├─────────────────────────────┬────────────────────────────────────┤
│ Queue (3):                  │ Commands:                           │
│   1. Song One              │   add     : Add music to queue      │
│   2. Song Two             │   pause   : Pause current track    │
│   3. Song Three           │   play    : Play/Resume track      │
├─────────────────────────────┴────────────────────────────────────┤
│ Last message here                                               │
└──────────────────────────────────────────────────────────────────┘
> Enter command: 
```

### Search Box

```
┌──────────────────────────────────────────────────────────────────┐
│                      SEARCH RESULTS                               │
├──────────────────────────────────────────────────────────────────┤
│   0: Song One                                                    │
│   1: Song Two                                                    │
│   2: Song Three                                                  │
└──────────────────────────────────────────────────────────────────┘
Select option: 
```

## Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `BoxWidth` | 103 | Total width of the display box |
| `TitleMaxLen` | 22 | Maximum length for title display |

## Lipgloss Styles

| Style | Usage |
|-------|-------|
| `titleStyle` | Main title "TOCADOR DE MUSICA" - bold, cyan, centered |
| `boxStyle` | Main box border - rounded, cyan border |
| `searchBoxStyle` | Search results box - rounded, blue border |
| `nowPlayingStyle` | Current track display - cyan text |
| `errorStyle` | Error messages - red text |
| `successStyle` | Success messages - green text |
| `inputStyle` | Input prompt styling - blue text |

## Struct: CLIinterface

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `reader` | `*bufio.Reader` | Stdin reader for input |
| `inputChan` | `chan string` | Channel for command input |
| `responseChan` | `chan string` | Channel for user responses |
| `waitingType` | `string` | Current waiting state ("input", "option", or "") |
| `wg` | `sync.WaitGroup` | Goroutine management |
| `mu` | `sync.Mutex` | Mutex for thread-safe access |
| `closed` | `bool` | Flag for closed state |
| `profileName` | `string` | Current profile name |
| `queue` | `[]string` | Current queue items |
| `nowPlaying` | `string` | Currently playing track |
| `isSearch` | `bool` | Search mode flag |
| `lastMessage` | `string` | Last system message |
| `isError` | `bool` | Error message flag |
| `isPaused` | `bool` | Paused state flag |

## Public Methods

### InputHandler Methods

| Method | Description |
|--------|-------------|
| `Input() <-chan string` | Returns input channel |
| `Run(ctx context.Context)` | Main input loop |
| `Close()` | Closes input channels |

### OutputHandler Methods

| Method | Description |
|--------|-------------|
| `Display(message string)` | Shows a message |
| `RequestInput(prompt string) <-chan string` | Requests user input |
| `DisplayOptions(options []string) <-chan int` | Shows options list |
| `FindUnknownCommand()` | Shows unknown command message |
| `ShowQueue(items []string)` | Displays queue |
| `ShowNowPlaying(track string)` | Shows current track |
| `Refresh()` | Refreshes display |
| `SetProfileName(name string)` | Sets profile name |

## Rendering

### renderBox()

Main display renderer that builds the complete UI:

1. **Header**: "TOCADOR DE MUSICA" centered
2. **Now Playing**: Track title in quotes or "paused"
3. **Queue | Commands**: Side-by-side using lipgloss table
4. **Message**: Last system message (error or success)
5. **Input**: Prompts outside the box

Key features:
- Single box wraps all content except input
- Queue shows max 10 items with "+N more" for overflow
- Queue titles truncated to fit with "..."
- Input prompt placed outside box for clarity

### buildQueueCommandsContent()

Builds the Queue|Commands section using lipgloss table:

- Queue column: Fixed width with truncation
- Commands column: Auto-sized based on command descriptions
- Uses `.Wrap(false)` to prevent text wrapping
- No outer borders (boxStyle provides them)

### renderSearch()

Search results display:

- Title: "SEARCH RESULTS"
- Options: Numbered list (0, 1, 2...)
- Select prompt: Outside the box

## Input Handling

- Reads from stdin using `bufio.Reader`
- Trims whitespace from input
- Non-blocking send to input channel
- Supports context cancellation
- Handles two modes:
  - **Command mode**: Sends to `inputChan`
  - **Response mode**: Sends to `responseChan` (for search selection)

## Thread Safety

- All public methods use mutex for state protection
- `RenderBox()` acquires lock before rendering
- Input/output operations are decoupled via channels

## Color Codes

| Color | Code | Usage |
|-------|------|-------|
| Cyan | 86 | Title, Now Playing |
| Blue | 75 | Search box border, Input |
| Red | 196 | Error messages |
| Green | 82 | Success messages |

## Notes

- The box is always 103 characters wide
- Queue items are limited to 10 with "+N more" indicator
- Input and "Select option:" prompts are OUTSIDE the box
- Uses lipgloss table for Queue|Commands alignment
- Truncation uses rune counting for proper UTF-8 handling
