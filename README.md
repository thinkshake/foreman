# 🏗️ foreman

**AI-native project management CLI for planning and preparing coding agent work.**

foreman is a file-based project management tool designed for AI assistants. It handles PM/EM responsibilities — requirements gathering, planning, high-level design, task breakdown into lanes, and progress tracking — to prepare work that coding agents can execute independently.

## Who is this for?

foreman is built for AI assistants (like those running on [OpenClaw](https://github.com/nichochar/openclaw)) that manage software projects. It bridges the gap between **planning** and **execution** by:

1. Structuring project knowledge (requirements, design, constraints)
2. Breaking work into **lanes** — self-contained workstreams
3. Generating **briefs** — compiled documents with all the context a coding agent needs

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

## Quick Start

```bash
# Initialize a project
foreman init --name my-project

# Set requirements
cat <<'EOF' | foreman req set
Build a REST API for managing tasks with:
- User authentication
- CRUD operations for tasks
- Real-time notifications via WebSocket
EOF

# Define the plan
cat <<'EOF' | foreman plan set
# Plan

## Goals
- Ship MVP in 2 weeks
- Support 1000 concurrent users

## Approach
- Go backend with chi router
- PostgreSQL for storage
- Redis for pub/sub
EOF

# Document design decisions
cat <<'EOF' | foreman design set
# Architecture

## Overview
Clean architecture with 3 layers:
- HTTP handlers (cmd/api)
- Business logic (internal/service)
- Data access (internal/repo)

## Database
PostgreSQL with migrations via golang-migrate.
EOF

# Break work into lanes
foreman lane add setup --summary "Project scaffolding, CI/CD, and dev tooling"
foreman lane add auth --summary "Authentication and authorization layer" --after setup
foreman lane add core-api --summary "CRUD endpoints for tasks" --after setup
foreman lane add realtime --summary "WebSocket notifications" --after core-api
foreman lane add deploy --summary "Production deployment and monitoring" --after core-api,auth

# Update lane statuses as work progresses
foreman lane status setup done
foreman lane status auth in-progress

# Check progress
foreman progress

# Generate a brief for a coding agent
foreman lane brief auth
```

## Commands

### Project Management

| Command | Description |
|---------|-------------|
| `foreman init [--name <name>] [--dir <path>]` | Initialize a new project |
| `foreman summary` | Show compact project summary |

### Requirements / Plan / Design

| Command | Description |
|---------|-------------|
| `foreman req show` | Print requirements |
| `foreman req set` | Set requirements (reads from stdin) |
| `foreman plan show` | Print plan |
| `foreman plan set` | Set plan (reads from stdin) |
| `foreman design show` | Print design document |
| `foreman design set` | Set design (reads from stdin) |

### Lane Management

| Command | Description |
|---------|-------------|
| `foreman lane add <name> [--summary "..."] [--after <deps>]` | Create a new lane |
| `foreman lane list [--status <status>]` | List all lanes |
| `foreman lane show <name>` | Print lane spec |
| `foreman lane set <name>` | Set lane spec (reads from stdin) |
| `foreman lane status <name> <status>` | Update lane status |
| `foreman lane brief <name>` | **Generate a coding agent brief** |

### Progress

| Command | Description |
|---------|-------------|
| `foreman progress` | Show progress dashboard |

### Lane Statuses

- `planned` — Defined but not started
- `ready` — Dependencies met, ready to start
- `in-progress` — Actively being worked on
- `review` — Work complete, under review
- `done` — Finished

## The Brief (Key Feature)

The `lane brief` command is foreman's primary output. It compiles everything a coding agent needs into a single, self-contained document:

```
# Lane Brief: auth

## Project Context
Name: my-project
Description: REST API for task management
Tech Stack: Go, PostgreSQL, Redis
Constraints: Must support 1000 concurrent users

## Requirements
Build a REST API for managing tasks with...

## Design Context
# Architecture
Clean architecture with 3 layers...

## Dependencies
- setup: `done` — Project scaffolding, CI/CD, and dev tooling

## Lane Spec
(full contents of the lane's markdown spec)

## Guidelines
- This lane is currently: in-progress
- Lane summary: Authentication and authorization layer

### Related Lanes
- setup (done): Project scaffolding, CI/CD, and dev tooling
- core-api (planned): CRUD endpoints for tasks
- realtime (planned): WebSocket notifications
- deploy (planned): Production deployment and monitoring
```

## File Structure

foreman manages a `.foreman/` directory in your project root:

```
.foreman/
├── project.yaml          # Project metadata, requirements, constraints
├── plan.md               # Planning overview (goals, approach, milestones)
├── design.md             # High-level architecture/design decisions
├── lanes/
│   ├── 01-setup.yaml     # Lane metadata (status, deps, summary)
│   ├── 01-setup.md       # Detailed spec for this lane
│   ├── 02-auth.yaml
│   ├── 02-auth.md
│   └── ...
└── progress.yaml         # Overall progress snapshot
```

### project.yaml

```yaml
name: "my-project"
description: "REST API for task management"
created: "2026-02-02T01:00:00+09:00"
updated: "2026-02-02T01:00:00+09:00"
requirements: |
  Build a REST API for managing tasks...
constraints: |
  Must support 1000 concurrent users
tech_stack:
  - "Go"
  - "PostgreSQL"
tags:
  - "backend"
  - "api"
```

### Lane YAML

```yaml
name: "auth"
order: 2
status: "in-progress"
summary: "Authentication and authorization layer"
dependencies:
  - "setup"
created: "2026-02-02T01:00:00+09:00"
updated: "2026-02-02T01:30:00+09:00"
```

## Behaviors

- **Auto-detection**: foreman finds `.foreman/` by walking up from the current directory (like `git` finds `.git/`)
- **Stdin piping**: All `set` commands read from stdin for easy scripting
- **Auto-numbering**: Lanes are automatically numbered (01-, 02-, 03-, etc.)
- **Progress tracking**: `progress.yaml` is automatically updated when lane statuses change
- **Colored output**: Status indicators use colors for quick visual scanning

## License

MIT — see [LICENSE](LICENSE).
