# 🏗️ foreman v3

**AI-native project management CLI for planning and preparing coding agent work.**

foreman is a file-based project management tool designed for AI assistants. It handles PM/EM responsibilities — requirements gathering, planning, high-level design, task breakdown into phases, and progress tracking — to prepare work that coding agents can execute independently.

## What's New in v3

- **Quick Mode** — `foreman quick "<task>"` skips design/phases for rapid builds
- **Workflow Presets** — `--preset nightly` or `--preset product` for common patterns
- **Auto-Advance Gates** — Confidence-based automatic gate approval
- **Progress Watching** — `foreman watch` for real-time progress monitoring

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

## Two Modes

### Quick Mode (v3)

For scripts, quick fixes, and nightly builds — skip the ceremony:

```bash
# Initialize and generate brief in one command
foreman quick "build a CLI that fetches weather data" --brief

# Or step by step:
foreman quick "build a weather CLI"
# Review .foreman/requirements.md
foreman gate requirements
foreman brief impl
```

Quick mode uses a streamlined workflow: **requirements → implementation**

### Full Mode

For serious products — full ceremony with design and phases:

```bash
# Initialize with full workflow
foreman init --name my-project
# Or with preset
foreman init --name my-project --preset product

# Define requirements
# Edit .foreman/requirements.md

# Pass requirements gate
foreman gate requirements

# Add design documents to .foreman/designs/
foreman gate design

# Create phases in .foreman/phases/
# e.g., phases/1-setup.md, phases/2-backend.md
foreman gate phases

# Generate briefs for coding agents
foreman brief 1-setup
foreman brief 2-backend

# Mark phases complete
foreman phase 1-setup done
foreman phase 2-backend done

# Complete implementation
foreman gate implementation
```

Full mode uses: **requirements → design → phases → implementation**

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `foreman init [--preset nightly\|product]` | Initialize a new project |
| `foreman quick "<task>" [--brief]` | Quick mode: skip design/phases |
| `foreman status` | Show project stage and gate status |
| `foreman gate [stage]` | Validate and control stage gates |
| `foreman brief <phase>` | Generate a coding agent brief |
| `foreman phase <name> <status>` | Update phase status |
| `foreman watch` | Watch project progress in real-time |

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

### Presets

| Preset | Mode | Auto-Advance | Reviewers |
|--------|------|--------------|-----------|
| `nightly` | Quick | 70% | All auto |
| `product` | Full | Off | Human for requirements/design |

```bash
# Quick builds for nightly development
foreman init --preset nightly
# Equivalent to: foreman init --quick

# Full workflow for products
foreman init --preset product
```

## The Brief (Key Feature)

The `brief` command compiles everything a coding agent needs:

**Quick mode** (`foreman brief impl`):
```markdown
# Implementation Brief

**Project:** weather-cli
**Mode:** Quick (no design/phases)

## Task
build a CLI that fetches weather data

## Requirements
(from .foreman/requirements.md)

## Implementation Guidelines
- Keep it simple and functional
- Write clean, readable code
- Include basic tests
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

## Dependencies
- ✅ 1-setup: done

## Phase Spec
(from .foreman/phases/2-backend.md)
```

## File Structure

### Quick Mode
```
.foreman/
├── config.yaml      # Project config (preset: nightly)
├── state.yaml       # Current stage, quick_mode: true
├── requirements.md  # Task description
└── briefs/
    └── impl.md      # Generated brief
```

### Full Mode
```
.foreman/
├── config.yaml      # Project config
├── state.yaml       # Current stage, gates, phases
├── requirements.md  # Project requirements
├── designs/         # Design documents
│   ├── architecture.md
│   └── api.md
├── phases/          # Phase specifications
│   ├── overview.md  # Phase overview
│   ├── 1-setup.md
│   └── 2-backend.md
└── briefs/          # Generated briefs
    ├── 1-setup.md
    └── 2-backend.md
```

## Progress Watching

Monitor project progress in real-time:

```bash
foreman watch

# Custom interval (default: 5s)
foreman watch --interval 10
```

Output:
```
👀 Watching project progress... (Ctrl+C to stop)

[01:23:45] Project: my-project
Mode: quick
Stages: ✅ requirements → 🔵 implementation

🎉 Stage advanced: requirements → implementation
   📝 Phase 1-setup: planned → in-progress
```

## Who is this for?

foreman is built for AI assistants (like those running on [OpenClaw](https://github.com/openclaw/openclaw)) that manage software projects. It bridges the gap between **planning** and **execution** by:

1. Structuring project knowledge (requirements, design, constraints)
2. Breaking work into **phases** — self-contained implementation units
3. Generating **briefs** — compiled documents with all the context a coding agent needs
4. **Quick mode** — skipping ceremony when speed matters

## License

MIT — see [LICENSE](LICENSE).
