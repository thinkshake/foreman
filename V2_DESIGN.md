# Foreman v2 — Design

## Philosophy

Foreman v2 is a **thin state gate + context compiler**. It enforces a structured development workflow by tracking which stage a project is at, preventing skipping ahead, and compiling rich briefs for coding agents. It does NOT do the actual work — AI agents (PM agents and coding agents) do.

## Core Concepts

### Stages

A project flows through ordered stages. Each stage has a **gate** that must be passed before the next stage unlocks.

Stages (in order):
1. **requirements** — Define what to build
2. **design** — Design the architecture
3. **phases** — Break implementation into phases
4. **implementation** — Build each phase (loops per phase)

### Gates

A gate validates that a stage is complete and controls advancement. Each gate has:
- **status**: `open` | `pending-review` | `approved` | `blocked`
  - `open` — stage is active, work can be done
  - `pending-review` — work done, waiting for review
  - `approved` — gate passed, next stage unlocked
  - `blocked` — previous stage not yet approved
- **reviewer**: `auto` | `human`
  - `auto` — the PM agent (Aston) reviews and approves
  - `human` — waits for human LGTM via `foreman gate <stage> --approve`

### Phases

Within the implementation stage, work is broken into numbered phases. Each phase can be independently briefed and handed to a coding agent. Phases have their own status tracking.

Phase statuses: `planned` | `in-progress` | `done`

## File Structure

```
.foreman/
├── config.yaml           # Project config + gate reviewer settings
├── state.yaml            # Current stage, gate statuses, phase statuses
├── requirements.md       # Requirements document (stage 1 output)
├── designs/              # Design documents (stage 2 output)
│   └── *.md
├── phases/               # Phase plans (stage 3 output)
│   ├── overview.md       # Phase breakdown overview
│   └── *.md              # Individual phase plans (1-setup.md, 2-backend.md, etc.)
└── briefs/               # Generated briefs (foreman brief output)
    └── *.md
```

### config.yaml

```yaml
name: "my-project"
description: "What this project does"
tech_stack:
  - TypeScript
  - React
  - SQLite
created: "2026-02-03T10:00:00+09:00"

# Gate reviewer configuration
reviewers:
  default: auto           # Default reviewer for all gates
  overrides:              # Per-stage overrides
    requirements: human   # Wait for human LGTM on requirements
    # design: auto        # (uses default)
    # phases: auto        # (uses default)
```

### state.yaml

```yaml
current_stage: design     # Which stage is currently active

gates:
  requirements:
    status: approved
    approved_at: "2026-02-03T10:30:00+09:00"
    approved_by: auto     # or "human"
  design:
    status: open
    approved_at: null
    approved_by: null
  phases:
    status: blocked
    approved_at: null
    approved_by: null
  implementation:
    status: blocked
    approved_at: null
    approved_by: null

# Phase tracking (populated after phases stage is approved)
phases:
  - name: "1-setup"
    status: planned       # planned | in-progress | done
  - name: "2-backend"
    status: planned
  - name: "3-frontend"
    status: planned
```

## CLI Commands

### `foreman init [--name <name>] [--dir <path>]`

Creates `.foreman/` directory with initial files:
- `config.yaml` with project name and default reviewers
- `state.yaml` with requirements gate `open`, all others `blocked`
- Empty `requirements.md` placeholder
- Empty `designs/`, `phases/`, `briefs/` directories

### `foreman status`

Shows the current project state:
```
Project: my-project
Stage: design (2/4)

Gates:
  ✅ requirements  approved (auto, 2026-02-03 10:30)
  🔵 design        open
  🔒 phases        blocked
  🔒 implementation blocked

Phases: (not yet defined)
```

When in implementation stage with phases:
```
Project: my-project
Stage: implementation (4/4)

Gates:
  ✅ requirements   approved
  ✅ design          approved
  ✅ phases          approved
  🔵 implementation  open

Phases:
  ✅ 1-setup      done
  🔵 2-backend    in-progress
  ⬜ 3-frontend   planned
```

### `foreman gate [<stage>]`

Without arguments: checks the current stage's gate readiness.
With stage name: checks that specific stage's gate.

**Validation logic:**
- `requirements`: checks `requirements.md` exists and is non-empty
- `design`: checks at least one `.md` file exists in `designs/`
- `phases`: checks `phases/overview.md` exists and at least one phase plan `.md`
- `implementation`: checks all phases are `done`

If validation passes:
- If reviewer is `auto`: automatically approves the gate, advances to next stage
- If reviewer is `human`: sets status to `pending-review`, prints message to wait for human

### `foreman gate <stage> --approve`

Human manually approves a gate. Only works when status is `pending-review`.
Advances to the next stage.

### `foreman gate <stage> --reject [--reason <text>]`

Rejects a gate review. Sets status back to `open` for rework.
Optional reason is stored in state.yaml.

### `foreman gate <stage> --reviewer auto|human`

Changes the reviewer for a specific stage gate.

### `foreman phase <name> <status>`

Updates a phase's status (`planned` | `in-progress` | `done`).
Only available during implementation stage.

### `foreman brief <phase-name>`

Compiles a self-contained brief for a coding agent. This is the killer feature.

The brief includes:
1. **Project context** — name, description, tech stack from config.yaml
2. **Requirements** — full requirements.md content
3. **Design context** — all design documents concatenated
4. **Phase overview** — the overall phase breakdown
5. **Phase spec** — the specific phase plan
6. **Dependencies** — status of preceding phases and their outputs
7. **Guidelines** — constraints, conventions, what to build

The brief is written to `.foreman/briefs/<phase-name>.md` and also printed to stdout.

### `foreman config show`

Prints current config.yaml.

### `foreman config set <key> <value>`

Updates a config value. Supports dot notation:
- `foreman config set name "my-project"`
- `foreman config set reviewers.default human`
- `foreman config set reviewers.overrides.requirements auto`

## Implementation Notes

### Keep from v1
- Go + cobra + yaml.v3 stack
- `project.FindRoot()` walk-up logic
- Brief compilation concept (but rewrite for phases)

### Remove from v1
- `req`, `plan`, `design` commands (just edit files directly)
- `lane` concept entirely (replaced by stages + phases)
- `progress` command (replaced by `status`)
- `summary` command (merged into `status`)

### New in v2
- `gate` command with validation + reviewer logic
- `state.yaml` with gate tracking
- `config.yaml` with reviewer settings
- `phase` command for implementation tracking
- `brief` rewritten for phase-based compilation
- `briefs/` output directory

### Testing
- Unit tests for gate validation logic
- Unit tests for state transitions (open → pending-review → approved)
- Unit tests for brief compilation
- Integration tests for full workflow: init → gate requirements → gate design → gate phases → brief → phase status
- Test edge cases: skipping stages, approving blocked gates, invalid transitions

### Error Handling
- Clear error messages when trying to skip stages
- Helpful suggestions: "Run `foreman gate requirements` to advance past requirements stage"
- Validate file existence before gate checks
