# Storage Shower Justfile

# Show available commands
default:
    @just --list

# Format all code
format:
    @echo "Formatting Go code..."
    go fmt ./...
    @echo "Formatting frontend code..."
    npx prettier --write "frontend/**/*.{js,html,css}"

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
    cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>storage-shower</string>
    <key>CFBundleIdentifier</key>
    <string>{{app_identifier}}</string>
    <key>CFBundleName</key>
    <string>{{app_name}}</string>
    <key>CFBundleDisplayName</key>
    <string>{{app_name}}</string>
    <key>CFBundleVersion</key>
    <string>{{app_version}}</string>
    <key>CFBundleShortVersionString</key>
    <string>{{app_version}}</string>
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
