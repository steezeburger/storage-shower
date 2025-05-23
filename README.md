# Storage Shower

A macOS disk space visualization tool that helps you understand where your storage is being used.

## Features

- Scan your file system to analyze disk usage
- Interactive treemap and sunburst visualizations
- Color coding by file type
- Detailed information for selected items
- Navigation through visualizations and breadcrumb trail
- Option to ignore hidden files

## Technology Stack

- **Backend**: Go for filesystem scanning and API
- **Frontend**: HTML, CSS, JavaScript with D3.js for visualizations
- **Integration**: webview for native macOS window

## Building the Application

### Prerequisites

- Go 1.16 or later
- Git

### Building from Source

1. Clone the repository
2. Install dependencies: `go mod download`
3. Build the application: `./build.sh`
4. The application will be available as `Storage Shower.app`

## Usage

1. Launch the application
2. Enter a path to scan or click "Home" to start from your home directory
3. Click "Scan" to begin analyzing disk usage
4. Once the scan completes, explore your disk usage through the interactive visualization
5. Click on directories to navigate deeper
6. Use the breadcrumb trail to navigate back up

## Implementation Details

### Core Components

- **main.go**: Core Go application with filesystem scanning and API
- **frontend/index.html**: HTML structure for the visualization UI
- **frontend/styles.css**: CSS styling for the application
- **frontend/app.js**: JavaScript for D3.js visualizations and UI interaction
- **build.sh**: Build script for packaging the macOS application

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
