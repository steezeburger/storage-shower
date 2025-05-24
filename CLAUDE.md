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
