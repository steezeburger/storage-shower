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
bundle:
    ./build.sh

# Clean build artifacts
clean:
    rm -rf storage-shower "Storage Shower.app"

# Install dependencies
deps:
    go mod download
    npm install