#!/bin/bash
set -e

# Build script for Storage Shower macOS application

# Configuration
APP_NAME="Storage Shower"
APP_BUNDLE="$APP_NAME.app"
APP_VERSION="1.0.0"
APP_IDENTIFIER="com.example.storageShower"

# Cleanup previous build
if [ -d "$APP_BUNDLE" ]; then
    echo "Removing previous build..."
    rm -rf "$APP_BUNDLE"
fi

# Create app bundle directory structure
echo "Creating app bundle structure..."
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

# Compile the Go application for macOS
echo "Compiling Storage Shower..."
GOOS=darwin GOARCH=amd64 go build -o "$APP_BUNDLE/Contents/MacOS/storage-shower" main.go

# Generate Info.plist
cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>storage-shower</string>
    <key>CFBundleIdentifier</key>
    <string>${APP_IDENTIFIER}</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleVersion</key>
    <string>${APP_VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${APP_VERSION}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2023</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.utilities</string>
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsLocalNetworking</key>
        <true/>
    </dict>
</dict>
</plist>
EOF

# Create a simple launcher script that wraps the binary
cat > "$APP_BUNDLE/Contents/MacOS/launcher" << EOF
#!/bin/bash
cd "\$(dirname "\$0")"
exec ./storage-shower
EOF

# Set permissions
echo "Setting permissions..."
chmod +x "$APP_BUNDLE/Contents/MacOS/storage-shower"
chmod +x "$APP_BUNDLE/Contents/MacOS/launcher"

# Update the CFBundleExecutable in Info.plist to point to our launcher
sed -i "" 's/<key>CFBundleExecutable<\/key>[ \t]*<string>storage-shower<\/string>/<key>CFBundleExecutable<\/key><string>launcher<\/string>/g' "$APP_BUNDLE/Contents/Info.plist"

# Create README file with instructions
cat > "README.md" << EOF
# Storage Shower

A macOS disk space visualization tool that helps you understand where your storage is being used.

## Requirements

- macOS 10.13 or later
- Permissions to access your file system

## Installation

Simply drag the \`${APP_NAME}.app\` to your Applications folder.

## Usage

1. Launch the application
2. Enter a path to scan or click "Home" to start from your home directory
3. Click "Scan" to begin analyzing disk usage
4. Once the scan completes, explore your disk usage through the interactive visualization
5. Click on directories to navigate deeper
6. Use the breadcrumb trail to navigate back up

## Features

- Interactive treemap and sunburst visualizations
- Detailed information about selected files and directories
- Color coding by file type
- Option to ignore hidden files
- Progress reporting during scans

## Building from Source

To build the application from source:

1. Ensure Go 1.16+ is installed
2. Install required dependencies: \`go get github.com/webview/webview\`
3. Run the build script: \`./build.sh\`

## License

This project is open source software.
EOF

echo "Build complete! The application is now available as ${APP_BUNDLE}"
echo ""
echo "To run the application, execute: open \"${APP_BUNDLE}\""
echo "To distribute the application, you can create a DMG or ZIP archive."
