#!/bin/bash

# Go development tools setup script
# Installs all the tools needed for development

set -e  # Exit on any error

echo "🚀 Setting up Go development tools..."
echo

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo "❌ Go is not installed. Please install Go first:"
    echo "   https://golang.org/doc/install"
    exit 1
fi

echo "📋 Go version: $(go version)"
echo "📁 GOPATH: $(go env GOPATH)"
echo

# Function to install a Go tool
install_tool() {
    local tool_name="$1"
    local import_path="$2"
    local description="$3"
    
    echo "📦 Installing $tool_name ($description)..."
    
    if go install "$import_path"; then
        echo "✅ $tool_name installed successfully"
    else
        echo "❌ Failed to install $tool_name"
        exit 1
    fi
    echo
}

# Install required tools
install_tool "goimports" "golang.org/x/tools/cmd/goimports@latest" "import organizer"
install_tool "staticcheck" "honnef.co/go/tools/cmd/staticcheck@latest" "static analyzer"
install_tool "deadcode" "golang.org/x/tools/cmd/deadcode@latest" "dead code detector"

# Check if tools are in PATH
echo "🔍 Verifying tool installation..."
echo

GOBIN_PATH="$(go env GOPATH)/bin"
if [[ ":$PATH:" != *":$GOBIN_PATH:"* ]]; then
    echo "⚠️  Warning: $GOBIN_PATH is not in your PATH"
    echo "   Add this to your shell profile (.bashrc, .zshrc, .config/fish/config.fish):"
    echo "   export PATH=\$PATH:\$(go env GOPATH)/bin"
    echo
fi

# Test each tool
test_tool() {
    local tool_name="$1"
    local test_command="$2"
    
    if command -v "$tool_name" >/dev/null 2>&1; then
        echo "✅ $tool_name is available"
    elif [ -x "$GOBIN_PATH/$tool_name" ]; then
        echo "✅ $tool_name is installed (at $GOBIN_PATH/$tool_name)"
    else
        echo "❌ $tool_name is not accessible"
    fi
}

test_tool "gofmt" "gofmt -h"
test_tool "goimports" "goimports -h"
test_tool "staticcheck" "staticcheck -version"
test_tool "deadcode" "deadcode -h"

echo
echo "🎉 Setup completed!"
echo
echo "📝 Next steps:"
echo "   1. Run './quality-check.sh' to check your code quality"
echo
echo "💡 Tip: Add the Go bin directory to your PATH if you see warnings above"
