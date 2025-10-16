# DocWorker: Multi-language Documentation IR & Pipeline (SQLite-first)

> CLI tool to extract *descriptive* documentation data from source code (no deep analysis), normalize it to a universal IR, and persist into SQLite per run.

---

## Output Layout (per run)
```
/out/<run-id>/
  manifest.json
  logs/
  artifacts/            # optional raw blobs (e.g., AST dumps)
  docdb.sqlite          # canonical IR database (this file is immutable after finalize)
  stats.json
```
- One SQLite per run (or shard by project if desired).

---

## SQLite Schema (v1)

> Naming is lower_snake_case. All times are UTC ISO-8601 strings. Large text columns use `TEXT`. Where noted, consider `FTS5` for search.

### project
| column           | type   | notes |
|------------------|--------|-------|
| id (PK)          | TEXT   | run-unique UUID |
| name             | TEXT   | logical project/repo name |
| root_uri         | TEXT   | filesystem or VCS root |
| tool_version     | TEXT   | semver of this tool |
| ir_schema        | TEXT   | e.g., `ir/1` |
| created_utc      | TEXT   | |

### container
Represents modules/packages/namespaces across languages.
| column                | type  | notes |
|-----------------------|-------|------|
| id (PK)               | TEXT  | stable within run |
| project_id (FK)       | TEXT  | -> project.id |
| language              | TEXT  | e.g., `go`, `python`, `csharp` |
| name                  | TEXT  | short name |
| full_name             | TEXT  | fully qualified (import path / module path) |
| kind                  | TEXT  | enum: `module|package|namespace` |
| version_tag           | TEXT  | optional (module version/commit) |
| doc_raw               | TEXT  | raw doc (container-level) |
| doc_fmt               | TEXT  | normalized/cleaned doc |
| extra_json            | TEXT  | language-specific data |

### file
| column           | type | notes |
|------------------|------|------|
| id (PK)          | TEXT | content digest-based id |
| project_id (FK)  | TEXT | -> project.id |
| path             | TEXT | relative path |
| digest           | TEXT | sha256 of content |
| language         | TEXT | |

### symbol
Represents classes/structs/interfaces/enums/funcs/methods/fields/consts/vars/type aliases/etc.
| column                | type  | notes |
|-----------------------|-------|------|
| id (PK)               | TEXT  | stable in run; hash of (language, container, kind, FQN, signature-hash) |
| container_id (FK)     | TEXT  | -> container.id |
| name                  | TEXT  | simple name |
| full_name             | TEXT  | fully-qualified |
| kind                  | TEXT  | enum, see **Symbol Kinds** |
| visibility            | TEXT  | `public|internal|private|package|protected` |
| flags                 | INTEGER | bitmask (abstract, static, generic, deprecated, etc.) |
| origin_file_id (FK)   | TEXT  | -> file.id |
| start_line            | INT   | |
| start_col             | INT   | |
| end_line              | INT   | |
| end_col               | INT   | |
| doc_raw               | TEXT  | original doc comment |
| doc_fmt               | TEXT  | cleaned/normalized doc |
| extra_json            | TEXT  | language-specific attributes |
| sid_hash              | TEXT  | cross-run stable hash (optional) |

**Indexes**
- `CREATE INDEX ix_symbol_name ON symbol(name COLLATE NOCASE);`
- Optional FTS: `symbol_doc_fts` over (`doc_raw`, `doc_fmt`).

### signature
One row per *callable* (functions/methods/constructors) and optionally for *types* (generic params).
| column             | type  | notes |
|--------------------|-------|------|
| symbol_id (PK, FK) | TEXT  | -> symbol.id |
| text               | TEXT  | pretty-printed signature |
| json               | TEXT  | machine form (see **Signature JSON**) |

### type_ref
Normalized type references owned by a symbol (params/results/fields/aliases).
| column            | type | notes |
|-------------------|------|------|
| id (PK)           | TEXT | |
| owner_symbol_id   | TEXT | -> symbol.id |
| slot              | TEXT | `param:<idx>`, `result:<idx>`, `field:<name>`, `alias`, `receiver`, `typeparam:<idx>` |
| json              | TEXT | **TypeRef JSON** |
| ord               | INT  | for stable ordering |

### member
Membership relations such as field-of, method-of, enum-value-of.
| column            | type | notes |
|-------------------|------|------|
| id (PK)           | TEXT | |
| owner_symbol_id   | TEXT | parent symbol |
| child_symbol_id   | TEXT | member symbol |
| ord               | INT  | declaration order |

### relation
Graph edges for navigation.
| column           | type  | notes |
|------------------|-------|------|
| src_symbol_id    | TEXT  | |
| rel              | TEXT  | enum, see **Relations** |
| dst_symbol_id    | TEXT  | |
| details_json     | TEXT  | e.g., reason, generic instantiation, notes |

(Compound index: `CREATE INDEX ix_relation ON relation(src_symbol_id, rel, dst_symbol_id);`)

### import
| column           | type | notes |
|------------------|------|------|
| container_id     | TEXT | -> container.id |
| target           | TEXT | raw import string |
| alias            | TEXT | alias if present |
| details_json     | TEXT | |

### diagnostic
| column       | type  | notes |
|--------------|-------|------|
| id (PK)      | TEXT  | |
| scope        | TEXT  | `project|container|file|symbol` |
| severity     | TEXT  | `info|warn|error` |
| code         | TEXT  | analyzer/loader code |
| message      | TEXT  | |
| file_id      | TEXT  | nullable |
| line         | INT   | |
| col          | INT   | |

---

## Universal Reference Types

### Symbol Kinds (enum)
`module, package, namespace, class, struct, interface, enum, union, typealias, function, method, constructor, field, property, const, var, parameter, result, receiver, generic_param, generic_arg`

### Relations (enum)
`declares, owns, overrides, implements, satisfies, embeds, mixes_in, imports, references, returns, parameter_of, receiver_of, type_alias_of, specializes`

### Signature JSON (callables)
```json
{
  "receiver": {"type": "pointer", "to": {"type": "named", "symbol": "pkg.Type"}},
  "type_params": [
    {"name": "T", "constraint": {"type": "interface", "methods": [], "embeds": []}}
  ],
  "params": [
    {"name": "ctx", "type": {"type": "named", "symbol": "context.Context"}, "variadic": false}
  ],
  "results": [
    {"name": "", "type": {"type": "named", "symbol": "error"}}
  ],
  "throws": []
}
```

### TypeRef JSON (sum type)
```json
{
  "type": "named" | "pointer" | "array" | "slice" | "map" | "chan" | "tuple" |
          "union" | "intersection" | "generic_inst" | "func" | "interface" | "struct" | "builtin",
  "name": "string",                 // for builtin or anonymous kinds
  "symbol": "FQN",                  // for named/generic references
  "args": [/* type refs */],        // for generic_inst
  "elem": {/* type ref */},         // for pointer/slice/chan
  "key": {/* type ref */},          // for map
  "fields": [{"name": "...", "type": {/* type ref */}, "embedded": false}], // struct/interface
  "tuple": [/* type refs */],       // tuple elements
  "direction": "send|recv|both"     // for channels (Go)
}
```

---

## Context Objects (essential)

### RunContext
- `RunID string`
- `OutDir string`
- `ToolVersion string`
- `IRSchema string`
- `Limits { Concurrency int; MemMB int }`
- `Logger Logger`
- `DB *sql.DB` (writer)
- `Manifest *Manifest`

### ProjectConfig
- `Name string`
- `RootURI string`
- `Languages []LanguageSpec` (enabled packs + settings)

### LanguageSpec
- `Language string` (`go`, `python`, ...)
- `PackVersion string`
- `Include/Exclude []string`
- `Build/Env map[string]string` (tags, OS/arch, etc.)

### PipelineContext (derived)
- `Run RunContext`
- `Lang LanguageSpec`
- `Container ContainerMeta` (during per-container execution)

### Artifacts (immutable)
- `Plan []ContainerMeta`
- `ContainerPlan { Files []FileMeta; BuildCfg map[string]string }`
- `ParsedUnitSet (opaque, language-native)`
- `IRFragment { Containers, Symbols, Signatures, TypeRefs, Members, Relations, Imports }`
- `PersistedStats`

---

## Pipeline Sketch (per language)

```text
[Plan] → [EnumerateFiles] → [ParseUnits] → [ExtractSymbols] → [Persist] → [Summarize]
```

- Steps are replaceable per language. Each step is a function: `func(ctx PipelineContext, in T) (U, error)`.
- Concurrency: container-level worker pool; transaction per container.

---

## Language Pack Contract

- `DiscoverContainers(root, spec) → []ContainerMeta`
- `EnumerateFiles(container, spec) → []FileMeta`
- `ParseUnits(files, build/env) → ParsedUnitSet`
- `ExtractSymbols(parsed) → IRFragment`
- `DocComments(unit) → normalized text + tags`
- `IntrinsicTypes() → map[string]BuiltInType]`
- Optional extras (language-specific relations/attributes).

Language packs produce **IR fragments** only; the core persists to SQLite and maintains indexes/FTS.

---

## Recommended Packages / Libraries

### CLI
- **spf13/cobra** — Commander for modern Go CLIs; supports nested commands, help/gen, autocompletion. (Docs: pkg.go.dev)  
  Ref: https://pkg.go.dev/github.com/spf13/cobra ; https://github.com/spf13/cobra
- **spf13/pflag** — GNU/POSIX-style flags, used by Cobra.  
  Ref: https://pkg.go.dev/github.com/spf13/pflag ; https://github.com/spf13/pflag
- **urfave/cli/v3** — Minimal, declarative alternative to Cobra.  
  Ref: https://pkg.go.dev/github.com/urfave/cli/v3 ; https://cli.urfave.org/

### SQLite (database/sql drivers)
- **github.com/mattn/go-sqlite3** — battle-tested CGO-backed driver.  
  Ref: https://pkg.go.dev/github.com/mattn/go-sqlite3 ; https://github.com/mattn/go-sqlite3
- **modernc.org/sqlite** — pure Go driver (no CGO), easier cross-compilation; good FTS5 support.  
  Ref: https://pkg.go.dev/modernc.org/sqlite

> Use SQLite **FTS5** for full-text search on docs.  
> Ref: https://www.sqlite.org/fts5.html

### Reading Go code (source → metadata)
- **golang.org/x/tools/go/packages** — load packages with build/tag awareness.  
  Ref: https://pkg.go.dev/golang.org/x/tools/go/packages
- **go/token, go/ast, go/types** — stdlib parsing & type information.  
  Ref: https://pkg.go.dev/golang.org/x/tools (module overview) and stdlib docs.

### Tree-sitter (multi-language parsing)
- **Official Go bindings**: `github.com/tree-sitter/go-tree-sitter`  
  Ref: https://pkg.go.dev/github.com/tree-sitter/go-tree-sitter ; https://github.com/tree-sitter/go-tree-sitter
- **Community bindings**: `github.com/smacker/go-tree-sitter` (w/ many grammars)  
  Ref: https://pkg.go.dev/github.com/smacker/go-tree-sitter ; https://github.com/smacker/go-tree-sitter
- Tree-sitter overview & language bindings: https://tree-sitter.github.io/tree-sitter/

> **Answer:** Yes—Go has Tree-sitter bindings (official and community).

---

## Default CLI Verbs

- `plan`    — enumerate containers; output summary
- `index`   — run pipeline; write `docdb.sqlite`
- `verify`  — integrity checks (FKs, counts, orphan checks)
- `search`  — quick FTS/name lookups
- `emit`    — export (JSONL/Graph) if needed

---

## Migrations & Versioning

- `PRAGMA user_version = <int>` for IR schema version.
- `manifest.json` tracks tool, IR schema, and language pack versions.
- Prefer additive migrations (new tables/cols) to preserve compatibility.

---

## Appendix: Minimal CREATE TABLEs

> Omitted constraints for brevity—add FKs and indexes as needed.

```sql
CREATE TABLE project (
  id TEXT PRIMARY KEY,
  name TEXT,
  root_uri TEXT,
  tool_version TEXT,
  ir_schema TEXT,
  created_utc TEXT
);

CREATE TABLE container (
  id TEXT PRIMARY KEY,
  project_id TEXT,
  language TEXT,
  name TEXT,
  full_name TEXT,
  kind TEXT,
  version_tag TEXT,
  doc_raw TEXT,
  doc_fmt TEXT,
  extra_json TEXT
);

CREATE TABLE file (
  id TEXT PRIMARY KEY,
  project_id TEXT,
  path TEXT,
  digest TEXT,
  language TEXT
);

CREATE TABLE symbol (
  id TEXT PRIMARY KEY,
  container_id TEXT,
  name TEXT,
  full_name TEXT,
  kind TEXT,
  visibility TEXT,
  flags INTEGER,
  origin_file_id TEXT,
  start_line INTEGER, start_col INTEGER,
  end_line INTEGER, end_col INTEGER,
  doc_raw TEXT, doc_fmt TEXT,
  extra_json TEXT,
  sid_hash TEXT
);

CREATE TABLE signature (
  symbol_id TEXT PRIMARY KEY,
  text TEXT,
  json TEXT
);

CREATE TABLE type_ref (
  id TEXT PRIMARY KEY,
  owner_symbol_id TEXT,
  slot TEXT,
  json TEXT,
  ord INTEGER
);

CREATE TABLE member (
  id TEXT PRIMARY KEY,
  owner_symbol_id TEXT,
  child_symbol_id TEXT,
  ord INTEGER
);

CREATE TABLE relation (
  src_symbol_id TEXT,
  rel TEXT,
  dst_symbol_id TEXT,
  details_json TEXT
);

CREATE TABLE import (
  container_id TEXT,
  target TEXT,
  alias TEXT,
  details_json TEXT
);

CREATE TABLE diagnostic (
  id TEXT PRIMARY KEY,
  scope TEXT,
  severity TEXT,
  code TEXT,
  message TEXT,
  file_id TEXT,
  line INTEGER,
  col INTEGER
);
```

---

## Tracking Checklist

- [ ] Initialize run directory & manifest
- [ ] Open SQLite and set `PRAGMA user_version`
- [ ] `plan`: discover containers (per language)
- [ ] `enumerate-files`: include/exclude/build cfg
- [ ] `parse`: Tree-sitter / language-native parsing
- [ ] `extract`: build IRFragment
- [ ] `persist`: transaction per container, bulk inserts
- [ ] `fts`: create/update FTS5 tables
- [ ] `verify`: counts, FK checks, orphan detection
- [ ] `summarize`: stats.json & finalize


---

## Concept: Containers vs. Project (And a Go `main` Example)

**Container** is the language-agnostic name for the top-level organizational unit that holds symbols. It maps to the idiomatic construct per language:

| Language | Typical Container | Kind |
|---|---|---|
| Go | Package (`main`, `net/http`) | `package` |
| Python | Module / package | `module` |
| C#/Java | Namespace / package | `namespace` / `package` |
| C/C++ | Translation unit or namespace | `file` / `namespace` |
| TS/JS | ES module | `module` |

A **Project** can have many **Containers**. Containers hold **Symbols** (types, functions, fields, etc.) and **Relations** (declares, calls, imports, implements, etc.).

### Example: Go `main` Function

Source:
```go
package main

import (
    "fmt"
    "os"
)

func main() {
    name := os.Getenv("USER")
    fmt.Println("Hello,", name)
}
```

**container** (row sketch)
| field | value |
|---|---|
| id | `container:go:main` |
| project_id | `project:myapp` |
| language | `go` |
| name | `main` |
| full_name | `main` |
| kind | `package` |

**symbol** (rows sketch)
| id | container_id | name | full_name | kind | visibility |
|---|---|---|---|---|---|
| `symbol:go:main.main` | `container:go:main` | `main` | `main.main` | `function` | `public` |
| `symbol:go:fmt` | `container:go:main` | `fmt` | `main.fmt` | `import` | `public` |
| `symbol:go:os`  | `container:go:main` | `os`  | `main.os`  | `import` | `public` |

**relation** (rows sketch)
| src_symbol_id | rel   | dst_symbol_id             | details_json |
|---|---|---|---|
| `symbol:go:main.main` | `calls` | `symbol:go:os.Getenv` | `{"arg":"USER"}` |
| `symbol:go:main.main` | `calls` | `symbol:go:fmt.Println` | `{"args":["Hello,","name"]}` |
| `symbol:go:main.main` | `references` | `symbol:go:os.Getenv` | `{}` |
| `symbol:go:main.main` | `references` | `symbol:go:fmt.Println` | `{}` |

**Notes for extraction (Go):**
- Collect `*ast.FuncDecl` for `main` and its position.
- Traverse `*ast.BlockStmt` with `ast.Inspect` to find `*ast.CallExpr`:
  - Callee from `SelectorExpr` (e.g., `fmt.Println`, `os.Getenv`).
  - Optional arg summaries for `details_json` (strings, identifiers) purely for doc navigation.

---

## Logging & TUI (Progress/Status)

### Goals
- Unified structured logging with human-friendly TUI.
- Live progress bars for pipeline stages and per-container work.
- Quiet/JSON mode for CI; interactive TUI for terminals.
- Deterministic, non-interfering with stdout/stderr when piping.

### Architecture
- **Logger** (structured): `slog` (Go 1.21+) or `zerolog`/`zap`.
- **Event bus** (optional): channel for progress events (`StartStep`, `Advance`, `Complete`, `Error`).
- **Renderer**: TUI that subscribes to events and renders bars/tables/spinners.
- **Bridging**: logger emits machine-readable JSON; TUI renders human view. Disable TUI when not a TTY.

### Suggested Libraries
- **Structured logging**
  - `log/slog` (stdlib) with custom handlers (JSON/Text).
  - `github.com/rs/zerolog/zerolog` (zero-allocation, JSON-first).
  - `go.uber.org/zap` (fast, widely used).
- **TUI / Progress**
  - Charmbracelet stack:
    - `github.com/charmbracelet/bubbletea` (state machine / reactive TUI)
    - `github.com/charmbracelet/lipgloss` (styling)
    - `github.com/charmbracelet/bubbles/progress` (progress bars)
    - `github.com/charmbracelet/bubbles/table`, `spinner`, `viewport` as needed
  - Alternative: `github.com/pterm/pterm` (quick progress bars/spinners) for simpler UIs.

### Event Model (sketch)
```
Event {
  RunID string
  Scope enum{Run, Container, File}
  Name string              // e.g., "ParseUnits"
  State enum{Start, Advance, Complete, Error}
  UnitID string            // container/file id
  Ordinal int              // step index
  Total int                // total steps/items (for progress)
  Value int                // current progress
  Msg string               // optional detail
  Time time.Time
}
```

### Rendering Ideas
- **Top bar**: run-wide progress (containers completed / total).
- **List**: top N active containers with per-step progress bars (`ParseUnits`, `ExtractSymbols`, `Persist`).
- **Panel**: recent warnings/errors with counts (click/keys to expand in TUI).
- **Footer**: hotkeys (q to quit, v to toggle verbose, j/k to navigate).

### Logger setup
- Config flags: `--log=json|text`, `--log-level=info|debug|warn|error`, `--no-tui`.
- When `stdout` is a TTY and `--log=text`, enable TUI and route detailed logs to a file (e.g., `logs/run.log`).
- In CI or when `--log=json`, disable TUI and send JSON logs to stdout.

### Progress semantics
- **Run-level**: total containers = N. Completed M → render `M/N` plus ETA (exponential smoothing).
- **Container-level**: fixed steps (Plan, Files, Parse, Extract, Persist). Each step can publish sub-item totals (e.g., files parsed).

### Error presentation
- Non-fatal errors increment a counter and appear in a collapsible panel.
- Fatal step error marks container as failed and continues to next container; diagnostic row written to DB.

### Minimal types (conceptual)
- `type ProgressScope int {Run, Container, File}`
- `type Step string`
- `type ProgressEvent struct { Scope, Step, State, UnitID, Ordinal, Total, Value, Msg, Time }`

### File outputs
- `logs/run.log` (full structured logs)
- `stats.json` (aggregate counts, durations, success/fail totals)
- Optional: `tui.recording.jsonl` with event stream for replay/regression UI tests.

