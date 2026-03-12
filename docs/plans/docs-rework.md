# Docs Structure Rework Plan

## Current state

```
docs/
├── cobracli/readme.md      # Dev note: how Cobra CLI was set up
├── edge-cloud.png          # Architecture image
├── file.sh                 # Stray shell script (does not belong here)
├── gitversion/readme.md    # Dev note: GitVersion setup
├── gpg-key/readme.md       # Dev note: GPG signing setup
├── init/readme.md          # Dev note: initial project bootstrap steps
├── lint/readme.md          # Dev note: linting setup
├── loadbalancer/readme.md  # Product doc: load balancer setup
├── openbao-migration.md    # Migration plan (flat file, inconsistent with rest)
├── readme.md               # Architecture overview (buried, not a README)
├── rke2/readme.md          # Product doc: RKE2 usage
├── test/readme.md          # Dev note: testing approach
├── todo-fixes.md           # This project's TODO fix plan
└── vault/readme.md         # Product doc: Vault/OpenBao integration
```

### Problems

1. **No navigation** — no root-level index linking all docs
2. **Dev notes mixed with product docs** — `cobracli/`, `gitversion/`, `gpg-key/`, `init/`, `lint/`, `test/` are internal developer setup notes, not useful to users or operators
3. **Inconsistent layout** — most docs live in a subdir with `readme.md`, but `openbao-migration.md` and `readme.md` are flat files
4. **`file.sh` in docs** — Proxmox VM creation script (`qm create`, `pvesh set`, disk resize). Personal infra scripting, completely unrelated to edgectl. Keep it.
5. **Single-file directories** — each dev note has its own folder just for one file; these should be consolidated
6. **Stale content** — `vault/readme.md` still references HashiCorp Vault; `readme.md` references HashiCorp Vault in the architecture section
7. **Plans living alongside docs** — `todo-fixes.md`, `openbao-migration.md` are project management artefacts, not reference docs
8. **Root `readme.md` is stale** — the project root `readme.md` (not in `docs/`) has significant overlap, stale TODOs ("TODO: Provide proper list of commands & purposes"), a "Done features" section with completed items, and references HashiCorp Vault while mentioning OpenBao as a future alternative. Not covered in the original rework plan.

---

## Proposed structure

```
docs/
├── README.md                   # Navigation index — links to all docs
├── architecture.md             # Promoted from docs/readme.md, updated for OpenBao
│
├── user/                       # End-user and operator docs
│   ├── getting-started.md      # Install + first-time setup (new)
│   ├── rke2.md                 # Moved + flattened from rke2/readme.md
│   ├── secret-management.md    # Rewritten from vault/readme.md (OpenBao)
│   └── loadbalancer.md         # Moved + flattened from loadbalancer/readme.md
│
├── development/                # Internal developer notes (consolidated)
│   ├── setup.md                # Merge of init/ + cobracli/ + gitversion/
│   ├── linting.md              # From lint/readme.md
│   ├── testing.md              # From test/readme.md
│   └── gpg-signing.md          # From gpg-key/readme.md
│
└── plans/                      # Project planning artefacts (not reference docs)
    ├── openbao-migration.md    # Moved from docs/openbao-migration.md
    └── todo-fixes.md           # Moved from docs/todo-fixes.md
```

---

## File-by-file migration

| Current path | Action | New path |
|---|---|---|
| `docs/readme.md` | Rename + update OpenBao refs | `docs/architecture.md` |
| `docs/rke2/readme.md` | Move + flatten | `docs/user/rke2.md` |
| `docs/vault/readme.md` | Rewrite for OpenBao | `docs/user/secret-management.md` |
| `docs/loadbalancer/readme.md` | Move + flatten | `docs/user/loadbalancer.md` |
| `docs/init/readme.md` | Merge into setup | `docs/development/setup.md` |
| `docs/cobracli/readme.md` | Merge into setup | `docs/development/setup.md` |
| `docs/gitversion/readme.md` | Merge into setup | `docs/development/setup.md` |
| `docs/lint/readme.md` | Move + flatten | `docs/development/linting.md` |
| `docs/test/readme.md` | Move + flatten | `docs/development/testing.md` |
| `docs/gpg-key/readme.md` | Move + flatten | `docs/development/gpg-signing.md` |
| `docs/openbao-migration.md` | Move | `docs/plans/openbao-migration.md` |
| `docs/todo-fixes.md` | Move | `docs/plans/todo-fixes.md` |
| `docs/edge-cloud.png` | Keep | `docs/edge-cloud.png` (referenced from architecture.md + root readme) |
| `docs/file.sh` | Delete — Proxmox VM script, unrelated to edgectl | — |
| `docs/docs-rework.md` | Move (this file) | `docs/plans/docs-rework.md` |

---

## `docs/README.md` outline

The new navigation index should cover:

```markdown
# EdgeCTL Documentation

## User Guide
- [Getting Started](user/getting-started.md)
- [RKE2 Cluster Management](user/rke2.md)
- [Secret Management (OpenBao)](user/secret-management.md)
- [Load Balancer Setup](user/loadbalancer.md)

## Reference
- [Architecture Overview](architecture.md)

## Development
- [Project Setup](development/setup.md)
- [Linting](development/linting.md)
- [Testing](development/testing.md)
- [GPG Signing](development/gpg-signing.md)

## Project Plans
- [OpenBao Migration](plans/openbao-migration.md)
- [TODO Fixes](plans/todo-fixes.md)
```

---

## Content updates needed alongside the restructure

| File | Update needed |
|---|---|
| `architecture.md` | Replace "HashiCorp Vault" with "OpenBao"; update Mermaid diagram label |
| `user/secret-management.md` | Full rewrite: OpenBao install, BAO_ADDR/BAO_TOKEN, bao CLI commands |
| `user/getting-started.md` | New file: prerequisites, install edgectl, set env vars, first cluster |
| `development/setup.md` | Merge three files; remove outdated steps (e.g. cobra-cli init is already done) |
| Root `readme.md` | Clean up stale TODOs, remove "Done features" section, update Secret Management section for OpenBao, write proper features list. **Keep it as the GitHub-facing README** — it should link to `docs/README.md` for full documentation. |

---

## Execution order

1. Create `docs/plans/` and move the plan files (openbao-migration.md, todo-fixes.md, docs-rework.md)
2. Create `docs/user/` — move and flatten the three product docs
3. Create `docs/development/` — consolidate the five dev note files
4. Delete now-empty directories: `cobracli/`, `gitversion/`, `gpg-key/`, `init/`, `lint/`, `test/`, `rke2/`, `loadbalancer/`, `vault/`
5. Rename `docs/readme.md` → `docs/architecture.md`, update content
6. Write `docs/README.md` navigation index
7. Write `docs/user/getting-started.md` (new content)
8. Clean up root `readme.md` — remove stale TODOs, "Done features", update for OpenBao, add link to `docs/README.md`
9.  Update `docs/user/secret-management.md` for OpenBao (can be done as part of or after the OpenBao migration)
