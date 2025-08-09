#!/bin/bash

# Go code quality check script
# Runs multiple Go tools to maintain code quality

set -e  # Exit on any error

echo "🔍 Running Go code quality checks..."
echo

# 1. Format code first (fixes spacing, alignment)
echo "📝 Formatting code with gofmt..."
gofmt -s -w .
echo "✅ gofmt completed"
echo

# 2. Organize imports (must come after gofmt to avoid conflicts)
echo "📦 Organizing imports with goimports..."
if command -v goimports >/dev/null 2>&1; then
    goimports -w .
    echo "✅ goimports completed"
else
    echo "⚠️ goimports not found, install with: go install golang.org/x/tools/cmd/goimports@latest"
fi
echo

# 3. Clean up dependencies (safe to run anytime)
echo "🧹 Cleaning up dependencies with go mod tidy..."
go mod tidy
echo "✅ go mod tidy completed"
echo

# 4. Run static analysis (after code is formatted and clean)
echo "🔬 Running static analysis with staticcheck..."
if command -v staticcheck >/dev/null 2>&1; then
    staticcheck ./...
    echo "✅ staticcheck completed - no issues found"
else
    echo "⚠️ staticcheck not found, install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
fi
echo

# 5. Check for dead code (run last since it's purely informational)
echo "💀 Checking for dead code..."
if command -v "$(go env GOPATH)/bin/deadcode" >/dev/null 2>&1; then
    DEAD_CODE_OUTPUT=$($(go env GOPATH)/bin/deadcode ./... 2>&1 || true)
    if [ -z "$DEAD_CODE_OUTPUT" ]; then
        echo "✅ No dead code found"
    else
        echo "⚠️ Dead code detected:"
        echo "$DEAD_CODE_OUTPUT"
    fi
else
    echo "⚠️ deadcode not found, install with: go install golang.org/x/tools/cmd/deadcode@latest"
fi
echo

# 6. Run tests
echo "🧪 Running tests..."
if go test ./test/...; then
    echo "✅ All tests passed"
else
    echo "❌ Tests failed"
    exit 1
fi
echo

# 7. Final build test
echo "🔨 Testing build..."
if go build -o gadget; then
    echo "✅ Build successful"
    rm -f gadget  # Clean up binary
else
    echo "❌ Build failed"
    exit 1
fi

echo
echo "🎉 All quality checks completed successfully!"
