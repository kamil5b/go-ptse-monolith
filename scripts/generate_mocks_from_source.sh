#!/usr/bin/env bash
set -uo pipefail

# Generate mocks using mockgen -source mode for each interface found under internal/
echo "Using mockgen from: $(go env GOPATH)/bin/mockgen"
PATH="$(go env GOPATH)/bin:$PATH"

count=0
processed_files=0
failed=0
while IFS= read -r -d $'\0' file; do
  # Skip mock files to avoid processing generated files
  if [[ "$file" =~ /mocks/ ]]; then
    continue
  fi
  
  # detect if file contains any interface declarations
  # This pattern catches 'type Name interface' anywhere in the file (not just start of line)
  iface_count=$(grep -o "type.*interface" "$file" 2>/dev/null | grep -v "TaskPayload\|Option" | wc -l)
  if [[ "$iface_count" -eq 0 ]]; then
    continue
  fi
  out_dir="$(dirname "$file")/mocks"
  mkdir -p "$out_dir"
  base=$(basename "$file" .go)
  out_file="$out_dir/mock_${base}.go"
  echo "Generating mocks from $file -> $out_file (interfaces: $iface_count)"
  if mockgen -source="$file" -destination="$out_file" -package="mocks" 2>&1; then
    ((count++)) || true
    ((processed_files++)) || true
  else
    echo "Warning: Failed to generate mocks for $file"
    ((failed++)) || true
  fi
done < <(find internal -type f -name '*.go' ! -name '*_test.go' -print0)

echo "Generated $count mock files from $processed_files source files."
if [[ "$failed" -gt 0 ]]; then
  echo "Failed to generate mocks for $failed files."
fi
