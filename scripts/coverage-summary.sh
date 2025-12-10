#!/bin/bash

# Check if coverage file exists
if [ ! -f coverage.out ]; then
    echo "❌ Coverage file not found. Run tests with coverage first."
    exit 1
fi

echo "📊 Filtering coverage data..."
# Filter out files we want to exclude from coverage
grep -v "internal/utils/tests.go" coverage.out | \
grep -v "prompts.go" | \
grep -v "mocks.go" > coverage.filtered.out

# Replace the original with filtered version
mv coverage.filtered.out coverage.out

echo "📊 Go Test Coverage Summary"
echo "================================"

# Generate detailed function-level coverage
go tool cover -func=coverage.out | sort -k3 -nr

echo ""
echo "📊 Package-level coverage breakdown:"
go tool cover -func=coverage.out | grep -v "total:" | awk '{
    package = $1;
    gsub(/\/[^\/]*:[^:]*$/, "", package);
    coverage_str = $3;
    gsub(/%/, "", coverage_str);
    coverage[package] += coverage_str + 0;
    count[package]++;
} END {
    for (pkg in coverage) {
        avg = coverage[pkg]/count[pkg];
        printf "%-60s (%5.1f%%)\n", pkg, avg
    }
}' | sort -k2 -nr

echo ""
echo "📊 Coverage data generated at coverage.out"

# Generate HTML coverage report
echo "📊 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html
echo "📊 HTML coverage report generated at coverage.html"
echo "💡 Excluded files: prompts.go, mocks.go, internal/utils/tests.go and files with //go:build !test tags"
