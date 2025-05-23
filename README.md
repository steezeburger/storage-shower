# Storage Shower

![Storage Shower Logo](images/logo_micro.png)

A disk space visualization tool that helps you understand where your storage is being used.

![Storage Shower Screenshot](images/screen1.png)

## Features

- Scan your file system to analyze disk usage
- Interactive treemap and sunburst visualizations
- Color coding by file type
- Detailed information for selected items
- Navigation through visualizations and breadcrumb trail
- Option to ignore hidden files
- Cancel scanning at any time
- Debugging mode for troubleshooting

## Technology Stack

- **Backend**: Go for filesystem scanning and API
- **Frontend**: HTML, CSS, JavaScript with D3.js for visualizations
- **Embedded Web**: Serves a web application locally in your browser

## Building the Application

### Prerequisites

- Go 1.16 or later
- Git
- Node.js and npm (for frontend formatting and linting)
- Just command runner (optional, for development tasks)

### Building from Source

1. Clone the repository
2. Install dependencies: `go mod download`
3. Run the application: `go run main.go`
4. For debugging, use: `go run main.go --debug`

### Code Formatting and Linting

The codebase uses automatic formatters and linters to maintain consistent code style and quality:

- **Go code**: 
  - Formatted with `go fmt`
  - Linted with `go vet`
- **Frontend code**: 
  - Formatted with Prettier (HTML, CSS, JavaScript)
  - JavaScript linted with ESLint
  - HTML linted with HTMLHint
  - CSS linted with Stylelint

To format all code, run:

```bash
just format
```

To lint all code, run:

```bash
just lint
```

To fix linting issues where possible, run:

```bash
just lint-fix
```

If you don't have Just installed, you can run the formatters directly:

```bash
# Format Go code
go fmt ./...

# Format frontend code
npx prettier --write "frontend/**/*.{js,html,css}"

# Lint Go code
go vet ./...

# Lint frontend code
npx eslint "frontend/**/*.js"
npx htmlhint "frontend/**/*.html"
npx stylelint "frontend/**/*.css"
```

Configuration files:
- `.prettierrc` - Configuration for Prettier
- `.editorconfig` - Editor configuration for consistent formatting
- `.eslintrc.json` - Configuration for ESLint
- `.htmlhintrc` - Configuration for HTMLHint
- `.stylelintrc.json` - Configuration for Stylelint
- `.golangci.yml` - Configuration for golangci-lint

## Usage

1. Launch the application
2. Enter a path to scan or click "Home" to start from your home directory
3. Click "Scan" to begin analyzing disk usage
4. While scanning, you can click "Stop" to cancel at any time
5. Once the scan completes, explore your disk usage through the interactive visualization
6. Click on directories to navigate deeper
7. Use the breadcrumb trail to navigate back up
8. Switch between treemap and sunburst visualizations as needed

## Development

### Code Formatting

We maintain consistent code style using automatic formatters:

- Go code is formatted with `go fmt`, the standard Go formatter
- Frontend code (HTML, CSS, JavaScript) is formatted with Prettier
- A `.editorconfig` file ensures consistent formatting across different editors

To format all code in the project, run:

```bash
just format
```

### Development Tasks

We use the Just command runner for common development tasks. Available tasks:

```bash
just                # Show available commands
just format         # Format all code
just lint           # Lint all code
just lint-fix       # Fix linting issues where possible
just run            # Run the application
just debug          # Run with debug mode
just build          # Build the application
just bundle         # Build macOS application bundle
just dmg            # Create DMG archive for distribution
just zip            # Create ZIP archive for distribution
just clean          # Clean build artifacts
just deps           # Install dependencies
just install-linters # Install linters
just validate       # Format and lint code
```

## Implementation Details

### Core Components

- **main.go**: Core Go application with filesystem scanning and API
- **frontend/index.html**: HTML structure for the visualization UI
- **frontend/styles.css**: CSS styling for the application
- **frontend/app.js**: JavaScript for D3.js visualizations and UI interaction

### Visualizations

- **Treemap**: Represent files and directories as nested rectangles
- **Sunburst**: Represent the file hierarchy as concentric rings

### Color Coding

- Directories: Blue
- Images: Red
- Videos: Purple
- Audio: Green
- Documents: Orange
- Archives: Yellow
- Other: Gray

## License

This project is open source software.

## Troubleshooting

If you encounter issues with file size calculations or scanning:

1. Run with debug mode: `go run main.go --debug`
2. Check the console output for detailed logging
3. For very large directories, scanning may take some time or stall on certain files
4. Use the Stop button to cancel a scan that's taking too long
