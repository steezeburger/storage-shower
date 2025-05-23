# Storage Shower Justfile

# Show available commands
default:
    @just --list

# Format all code
format *args="":
    @echo "Formatting Go code..."
    go fmt ./...
    @echo "Formatting frontend code..."
    @if args == "--check"
    npx prettier --check "frontend/**/*.{js,html,css}"
    @else
    npx prettier --write "frontend/**/*.{js,html,css}"
    @endif

# Lint all code
lint: lint-go lint-js lint-html lint-css

# Lint Go code
lint-go:
    @echo "Linting Go code..."
    go vet ./...

# Lint JavaScript
lint-js:
    @echo "Linting JavaScript..."
    npx eslint "frontend/**/*.js"

# Lint HTML
lint-html:
    @echo "Linting HTML..."
    npx htmlhint "frontend/**/*.html"

# Lint CSS
lint-css:
    @echo "Linting CSS..."
    npx stylelint "frontend/**/*.css"

# Fix linting issues where possible
lint-fix:
    @echo "Fixing JavaScript linting issues..."
    npx eslint --fix "frontend/**/*.js"
    @echo "Fixing CSS linting issues..."
    npx stylelint --fix "frontend/**/*.css"

# Run the application
run *args:
    go run main.go {{args}}

# Run with debug mode
debug:
    just run --debug

# Build the application
build:
    go build -o storage-shower main.go

# Build macOS application bundle
bundle app_name="Storage Shower" app_version="1.0.0" app_identifier="com.example.storageShower":
    #!/bin/bash
    set -e

    # Configuration
    APP_BUNDLE="{{app_name}}.app"

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
    echo '<?xml version="1.0" encoding="UTF-8"?>' > "$APP_BUNDLE/Contents/Info.plist"
    echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '<plist version="1.0">' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '<dict>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleExecutable</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>storage-shower</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleIdentifier</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>{{app_identifier}}</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleName</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>{{app_name}}</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleDisplayName</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>{{app_name}}</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleVersion</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>{{app_version}}</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleShortVersionString</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>{{app_version}}</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundleInfoDictionaryVersion</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>6.0</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>CFBundlePackageType</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>APPL</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>NSHighResolutionCapable</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <true/>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>NSHumanReadableCopyright</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>Copyright © 2023</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>LSMinimumSystemVersion</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>10.13</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>LSApplicationCategoryType</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <string>public.app-category.utilities</string>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <key>NSAppTransportSecurity</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    <dict>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '        <key>NSAllowsLocalNetworking</key>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '        <true/>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '    </dict>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '</dict>' >> "$APP_BUNDLE/Contents/Info.plist"
    echo '</plist>' >> "$APP_BUNDLE/Contents/Info.plist"

    # Create a simple launcher script that wraps the binary
    echo '#!/bin/bash' > "$APP_BUNDLE/Contents/MacOS/launcher"
    echo 'cd "$(dirname "$0")"' >> "$APP_BUNDLE/Contents/MacOS/launcher"
    echo 'exec ./storage-shower' >> "$APP_BUNDLE/Contents/MacOS/launcher"

    # Set permissions
    echo "Setting permissions..."
    chmod +x "$APP_BUNDLE/Contents/MacOS/storage-shower"
    chmod +x "$APP_BUNDLE/Contents/MacOS/launcher"

    # Update the CFBundleExecutable in Info.plist to point to our launcher
    sed -i "" 's/<key>CFBundleExecutable<\/key>[ \t]*<string>storage-shower<\/string>/<key>CFBundleExecutable<\/key><string>launcher<\/string>/g' "$APP_BUNDLE/Contents/Info.plist"

    echo "Build complete! The application is now available as $APP_BUNDLE"
    echo ""
    echo "To run the application, execute: open \"$APP_BUNDLE\""
    echo "To distribute the application, you can create a DMG or ZIP archive."

# Clean build artifacts
clean:
    rm -rf storage-shower "Storage Shower.app"

# Install dependencies
deps:
    go mod download
    npm install

# Install linters
install-linters:
    @echo "Installing Go linters..."
    go install golang.org/x/lint/golint@latest
    @echo "Installing frontend linters (via npm)..."
    npm install --save-dev eslint htmlhint stylelint stylelint-config-standard

# Run backend tests
test-backend:
    go test -v ./...

# Run frontend tests
test-frontend:
    npx jest --config frontend/jest.config.js

# Run all tests
test: test-backend test-frontend

# Validate code (format and lint)
validate: format lint

# Create a DMG file for distribution
dmg app_name="Storage Shower": bundle
    @echo "Creating DMG file..."
    hdiutil create -volname "{{app_name}}" -srcfolder "{{app_name}}.app" -ov -format UDZO "{{app_name}}.dmg"
    @echo "DMG created: {{app_name}}.dmg"

# Create a ZIP file for distribution
zip app_name="Storage Shower": bundle
    @echo "Creating ZIP file..."
    zip -r "{{app_name}}.zip" "{{app_name}}.app"
    @echo "ZIP created: {{app_name}}.zip"
