#!/bin/bash

HTML_FILE="coverage.html"
MIN_COVERAGE=80

if [ ! -f "$HTML_FILE" ]; then
    echo "❌ HTML coverage file not found: $HTML_FILE"
    exit 1
fi

echo "Checking coverage requirements..."

# Extract only <option> lines and get file + percentage
failed_files=$(grep "<option value=" "$HTML_FILE" \
  | sed -E 's/.*>([^<]+) \(([0-9]+\.[0-9]+)%\)<\/option>/\1 \2/' \
  | awk -v min="$MIN_COVERAGE" '
    {
      file = $1
      for (i=2; i<NF; i++) file = file " " $i
      perc = $NF
      if (perc+0 < min) print file " (" perc "%)"
    }'
)

# Final result
if [ -n "$failed_files" ]; then
    echo "❌ Files below ${MIN_COVERAGE}% coverage:"
    echo "$failed_files"
    exit 1
else
    echo "✅ All files meet ${MIN_COVERAGE}% coverage requirement"
fi
