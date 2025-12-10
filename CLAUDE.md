# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform Provider for Algolia - enables managing Algolia resources (indexes, API keys, rules, synonyms, query suggestions) through Terraform.

## Common Commands

```bash
# Build and install provider for local development (version 999.999.999)
make install

# Run all acceptance tests (requires ALGOLIA_APP_ID and ALGOLIA_API_KEY env vars)
make testacc

# Run a specific test
TF_ACC=1 go test ./internal/provider -v -run TestAccResourceIndex

# Generate documentation (run after schema or examples changes)
make generate

# Lint code
make lint

# Format code (Go + Terraform)
make fmt

# Clean up test resources
make teardown
```

## Environment Setup

Tests require Algolia credentials:
```bash
export ALGOLIA_APP_ID=<your-app-id>
export ALGOLIA_API_KEY=<your-api-key>
```

Or use direnv with `.env` file.

## Architecture

### Provider Structure

- `main.go` - Entry point, registers provider with Terraform plugin SDK v2
- `internal/provider/` - Core provider implementation
  - `provider.go` - Provider configuration, creates Algolia API clients
  - `resource_*.go` - Resource implementations (index, virtual_index, api_key, rule, synonyms, query_suggestions)
  - `data_source_*.go` - Data source implementations
  - `*_test.go` - Acceptance tests for each resource/data source
- `internal/algoliautil/` - Algolia API utilities (error handling, HTTP debugging, region mapping)
- `internal/mutex/` - Key/Value mutex for concurrent resource operations

### Supported Resources

- `algolia_index` - Manages index settings (primary and standard replicas)
- `algolia_virtual_index` - Manages virtual replica indexes
- `algolia_api_key` - Manages API keys with ACL permissions
- `algolia_rule` - Manages index rules
- `algolia_synonyms` - Manages synonym sets
- `algolia_query_suggestions` - Manages query suggestions configuration

### Data Sources

- `algolia_index` - Reads existing index settings
- `algolia_virtual_index` - Reads existing virtual index settings

### Documentation Generation

Documentation is auto-generated from schema descriptions and `examples/` directory:
- `docs/` - Generated documentation (do not edit manually)
- `templates/` - Documentation templates
- `examples/` - Example Terraform configurations used in docs

Run `make generate` to regenerate docs after schema or example changes.

## Testing Notes

- Acceptance tests create real Algolia resources and may incur costs
- Tests are run against Terraform versions 1.3.9, 1.4.6, and 1.5.1 in CI
- Use `TF_LOG=DEBUG` for detailed provider logging