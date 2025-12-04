# Generating mocks

This repo includes a small helper script to generate mocks for interfaces using `mockgen` from `github.com/golang/mock`.

Script: `scripts/generate_mocks.sh`

Quick steps

1. Install mockgen:

   go install github.com/golang/mock/mockgen@latest

2. Create a spec file (for example `mocks.spec`) with lines like:

   github.com/kamil5b/github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain ProductRepository internal/modules/product/repository/mock_product_repo.go mocks

3. Run the generator:

   ./scripts/generate_mocks.sh mocks.spec

Spec file format (one entry per non-empty, non-comment line):

  <package> <interface> <output_file> [output_package]

Where:
  - `package` is the import path containing the interface
  - `interface` is the interface name to mock
  - `output_file` is the file to write the mock to (use '-' for stdout)
  - `output_package` is optional package name for the generated mock (defaults to `mocks`)

Notes
- The script uses reflect mode (mockgen <package> <interface>), so the target package must be buildable.
- If directories do not exist they will be created.
