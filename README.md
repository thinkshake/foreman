# 🏗️ foreman v2.1

**AI-native project management CLI for planning and preparing coding agent work.**

foreman is a file-based project management tool designed for AI assistants. It handles PM/EM responsibilities — requirements gathering, planning, high-level design, task breakdown into phases, and progress tracking — to prepare work that coding agents can execute independently.

## What's New in v2.1

- **Workflow Presets** — `minimal`, `light`, `full` for different project sizes
- **TDD Integration** — `--tdd` flag enables test-driven development mode with brief injection
- **Custom Workflows** — Power users can define custom workflow stages
- **Backward Compatible** — `nightly` and `product` presets still work (mapped to `minimal` and `full`)

## Installation

```bash
go install github.com/thinkshake/foreman@latest
```

Or build from source:

```bash
git clone https://github.com/thinkshake/foreman.git
cd foreman
go build -o foreman .
```

## Workflow Presets

| Preset | Use Case | Gates | Stages | Auto-Advance |
|--------|----------|-------|--------|--------------|
| `minimal` | Script, hotfix | 0 | requirements → implementation | 100% |
| `light` | Small tool, feature | 1 | requirements → implementation | 70% |
| `full` | Product, complex system | 3 | requirements → design → phases → implementation | Off |

### Minimal Mode

For scripts, hotfixes, and quick builds — no gates, straight to implementation:

```bash
foreman init --preset minimal --name my-script
# Requirements auto-approved, already at implementation
foreman brief impl
```

### Light Mode

For small tools and features — one gate, no design phase:

```bash
foreman init --preset light --name my-tool
# Edit .foreman/requirements.md
foreman gate requirements   # One gate to pass
foreman brief impl
```

### Full Mode

For products and complex systems — full ceremony with design and phases:

```bash
foreman init --preset full --name my-product
# Edit .foreman/requirements.md
foreman gate requirements
# Add designs to .foreman/designs/
foreman gate design
# Add phases to .foreman/phases/
foreman gate phases
foreman brief 1-setup
foreman brief 2-backend
```

## TDD Integration

Enable test-driven development with the `--tdd` flag:

```bash
foreman init --preset light --tdd --name my-project
```

This adds a `testing` block to your config:

```yaml
testing:
  style: tdd        # tdd | coverage | none
  required: false   # Block phase completion without tests?
  framework: vitest # Hint for coding agent
```

When TDD is enabled, `foreman brief` injects TDD instructions:

```markdown
## Test-Driven Development

⚠️ **TDD is enabled for this project.** Follow this workflow:

1. **Write tests first** — Define expected behavior before implementation
2. **Run tests (they should fail)** — Confirm the test is valid
3. **Implement the feature** — Write minimal code to pass the test
4. **Refactor** — Clean up while keeping tests green
5. **Repeat** — For each feature/function
```

## Custom Workflows

Power users can define custom workflow stages in config.yaml:

```yaml
workflow:
  - requirements
  - implementation
```

Valid stages: `requirements`, `design`, `phases`, `implementation`

The workflow must include at least `implementation`.

## Quick Mode (Legacy v3)

The `foreman quick` command from v3 is still supported:

```bash
foreman quick "build a CLI that fetches weather data" --brief
```

This is equivalent to minimal preset with inline task specification.

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `foreman init [--preset minimal\|light\|full] [--tdd]` | Initialize a new project |
| `foreman quick "<task>" [--brief]` | Quick mode: skip design/phases |
| `foreman status` | Show project stage and gate status |
| `foreman gate [stage]` | Validate and control stage gates |
| `foreman brief <phase>` | Generate a coding agent brief |
| `foreman phase <name> <status>` | Update phase status |
| `foreman watch` | Watch project progress in real-time |

### Preset Aliases (Backward Compat)

```bash
foreman init --preset nightly   # → minimal
foreman init --preset product   # → full
foreman init --quick            # → minimal
```

### Gate Operations

```bash
# Check current gate
foreman gate

# Check specific gate
foreman gate requirements
foreman gate design

# Manual approval (for human-reviewed gates)
foreman gate requirements --approve
foreman gate requirements --reject --reason "Missing acceptance criteria"

# Set reviewer type
foreman gate requirements --reviewer human
foreman gate requirements --reviewer auto
```

## The Brief (Key Feature)

The `brief` command compiles everything a coding agent needs:

**Minimal/Light mode** (`foreman brief impl`):
```markdown
# Implementation Brief

**Project:** my-tool
**Mode:** Light (requirements gate only)
**Testing:** TDD enabled

## Task
build a CLI that fetches weather data

## Requirements
(from .foreman/requirements.md)

## Test-Driven Development
(TDD workflow instructions if enabled)

## Implementation Guidelines
- Keep it simple and functional
- Write clean, readable code
- Write tests first (TDD enabled)
- Add a README with usage instructions
```

**Full mode** (`foreman brief 2-backend`):
```markdown
# Phase Brief: 2-backend

**Status:** planned

## Project Context
Name: my-project
Tech Stack: Go, PostgreSQL

## Requirements
(from .foreman/requirements.md)

## Design Context
(from .foreman/designs/*.md)

## Test-Driven Development
(TDD workflow instructions if enabled)

## Phase Spec
(from .foreman/phases/2-backend.md)
```

## File Structure

### Minimal/Light Mode
```
.foreman/
├── config.yaml      # Project config (preset: minimal|light)
├── state.yaml       # Current stage, workflow
├── requirements.md  # Task description
└── briefs/
    └── impl.md      # Generated brief
```

### Full Mode
```
.foreman/
├── config.yaml      # Project config (preset: full)
├── state.yaml       # Current stage, gates, phases
├── requirements.md  # Project requirements
├── designs/         # Design documents
│   ├── architecture.md
│   └── api.md
├── phases/          # Phase specifications
│   ├── overview.md
│   ├── 1-setup.md
│   └── 2-backend.md
└── briefs/          # Generated briefs
    ├── 1-setup.md
    └── 2-backend.md
```

## Config Schema (v2.1)

```yaml
name: my-project
description: ""
tech_stack:
  - Go
  - PostgreSQL
created: 2026-02-19T00:00:00Z
reviewers:
  default: auto          # auto | human
  overrides:
    requirements: human  # Per-stage override
preset: light            # minimal | light | full
auto_advance: 70         # 0-100 confidence threshold
testing:                 # TDD configuration
  style: tdd             # tdd | coverage | none
  required: false        # Block without tests?
  framework: vitest      # Hint for coding agent
  min_cover: 80          # Minimum coverage (coverage style)
workflow:                # Custom workflow (optional)
  - requirements
  - implementation
```

## Progress Watching

Monitor project progress in real-time:

```bash
foreman watch

# Custom interval (default: 5s)
foreman watch --interval 10
```

## Who is this for?

foreman is built for AI assistants (like those running on [OpenClaw](https://github.com/openclaw/openclaw)) that manage software projects. It bridges the gap between **planning** and **execution** by:

1. **Structuring project knowledge** — requirements, design, constraints
2. **Breaking work into phases** — self-contained implementation units
3. **Generating briefs** — compiled documents with all the context a coding agent needs
4. **Supporting different workflows** — minimal for scripts, full for products
5. **Enforcing TDD** — optional test-driven development mode

## License

MIT — see [LICENSE](LICENSE).
