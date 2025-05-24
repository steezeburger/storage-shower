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

## Automated Browser Testing with SSE Server + Playwright MCP

This project supports automated browser testing using the SSE server with Playwright MCP. This allows for comprehensive testing of UI changes and user interactions.

### Basic Testing Workflow:

1. **Kill existing processes**: `pkill -f "storage-shower\|go run" 2>/dev/null || true`
2. **Build and start server**: For embedded file changes (HTML/CSS/JS), rebuild first:
   ```bash
   go build -o storage-shower . && nohup ./storage-shower > /dev/null 2>&1 & echo "Server started"
   ```
   For Go-only changes, use: `nohup just run > /dev/null 2>&1 & echo "Server started"`
3. **Test with browser automation**: Use the SSE server browser tools to navigate, interact, and verify changes

### Important Notes for Embedded Files:

- **Web files (HTML/CSS/JS) are embedded at compile time** using Go's `//go:embed` directive
- Changes to `web/` files require rebuilding the binary to take effect
- Use `go build` followed by running the binary for embedded file changes
- Use `just run` (go run) only when Go source files change

### Browser Testing Commands:

```bash
# Navigate to the application
mcp__sse-server__browser_navigate(url="http://localhost:8080")

# Take screenshots to verify UI
mcp__sse-server__browser_take_screenshot()

# Interact with elements (example: start a scan)
mcp__sse-server__browser_type(element="textbox", text="/path/to/test")
mcp__sse-server__browser_click(element="button 'Scan'")

# Wait for operations to complete
mcp__sse-server__browser_wait_for(time=3)

# Verify results with snapshots
mcp__sse-server__browser_snapshot()

# Check for JavaScript errors and network issues
mcp__sse-server__browser_console_messages()  # Check for JS errors, 404s, etc.
mcp__sse-server__browser_network_requests()  # View all network requests and responses
```

### Complete Testing Example:

```bash
# 1. Stop existing server
pkill -f "storage-shower\|go run" 2>/dev/null || true

# 2. Rebuild server with updated embedded files (if web files changed)
go build -o storage-shower . && nohup ./storage-shower > /dev/null 2>&1 & echo "Server rebuilt and started"

# 3. Test the changes using browser automation
# - Navigate to http://localhost:8080
# - Perform user interactions
# - Take screenshots to verify UI
# - Test functionality with scans
# - Verify new features work as expected
```
