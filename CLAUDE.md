# CLAUDE.md - Guidelines for Storage Shower Project

## Project Overview

Storage Shower is a disk space visualization tool that helps users understand where their storage is being used. It provides interactive treemap and sunburst visualizations of disk usage, with detailed information about files and directories.

## Code Structure

- `main.go` - Backend Go server with API endpoints
- `web/` - Web application
  - `index.html` - Main HTML structure
  - `app.js` - Frontend JavaScript application
  - `styles.css` - CSS styling

## Coding Standards

### Go

- Use Go standard formatting (gofmt)
- Prefer explicit error handling over panic
- Functions should have clear documentation comments
- Maintain idiomatic Go code style
- Use proper error handling and logging

### JavaScript

- Use ES6+ features where appropriate
- Prefer const/let over var
- Use clear, descriptive variable and function names
- Add comments for complex logic
- Use consistent indentation (2 spaces)

### CSS

- Use descriptive class names
- Organize styles logically
- Use CSS variables for theme colors and repeated values
- Ensure responsive design considerations

## Pull Request Guidelines

When creating pull requests:

1. Include clear descriptions of changes
2. Reference any related issues
3. Ensure code passes existing functionality
4. Include tests for new features when possible
5. Update documentation as needed

## Project-Specific Considerations

- User experience is a priority - focus on intuitive interfaces
- Performance is critical for large directory scans
- Handle edge cases like access permissions gracefully
- Maintain responsive UI even during scanning operations
- Provide clear error messages to users

## Security Considerations

- Validate user input (especially file paths)
- Handle file system errors appropriately
- Do not expose sensitive system information
- Consider permissions when scanning directories

## Feature Implementation Guidance

When implementing new features:

1. Maintain the existing architecture separation (Go backend, JS web)
2. Use the existing event-based communication patterns
3. Follow the established UI/UX patterns
4. Consider both performance and usability

## Just Commands

This project uses [just](https://github.com/casey/just) as a command runner. Available commands:

### Development
- `just` or `just --list` - Show all available commands
- `just run [args]` - Run the application
- `just debug` - Run with debug mode
- `just deps` - Install dependencies (Go modules + npm packages)

### Building
- `just build` - Build the Go binary
- `just bundle [app_name] [app_version] [app_identifier]` - Build macOS app bundle
- `just dmg [app_name]` - Create DMG for distribution (builds bundle first)
- `just zip [app_name]` - Create ZIP for distribution (builds bundle first)

### Code Quality
- `just format` or `just f` - Format all code (Go + web)
- `just format --check` - Check formatting without making changes
- `just lint` or `just l` - Lint all code (Go, JS, HTML, CSS)
- `just lint-go` - Lint only Go code
- `just lint-js` - Lint only JavaScript
- `just lint-html` - Lint only HTML
- `just lint-css` - Lint only CSS
- `just lint-fix` - Fix linting issues where possible
- `just validate` - Format and lint all code

### Testing
- `just test` or `just t` - Run all tests (backend + web)
- `just test-backend` - Run only Go tests
- `just test-web` - Run only web tests

### Utilities
- `just clean` - Remove build artifacts
- `just install-linters` - Install required linting tools
- `just prepush` - Run all checks before pushing (format, test, lint Go)

# Testing Feedback Loop Workflow

When making changes to the web frontend and testing them:

1. **Start the server efficiently**: Use `nohup just run > /dev/null 2>&1 & echo "Server started"` to avoid timeout issues
2. **Kill existing processes if needed**: Use `pkill -f storage-shower; pkill -f "go run"` to stop any running servers
3. **After making JavaScript/CSS changes**: Always refresh the browser (`browser_navigate` to the same URL) to pick up the latest changes
4. **For Go backend changes**: Restart the server completely since Go changes require recompilation

Example workflow:
```bash
# Kill any existing server
pkill -f storage-shower; pkill -f "go run"

# Start server in background 
nohup just run > /dev/null 2>&1 & echo "Server started"

# Make your changes to web/app.js or web/styles.css

# Refresh browser to test changes
# (browser_navigate to http://localhost:8080)
```
