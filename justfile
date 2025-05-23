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
bundle:
    ./build.sh

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
