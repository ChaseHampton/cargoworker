PRAGMA journal_mode = WAL;
PRAGMA page_size = 4096;
PRAGMA auto_vacuum = INCREMENTAL;
VACUUM;

PRAGMA foreign_keys = ON;

CREATE TABLE project (
  id           INTEGER PRIMARY KEY,
  name         TEXT NOT NULL,
  root_path    TEXT NOT NULL,
  created_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE container (
  id           INTEGER PRIMARY KEY,
  project_id   INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
  module_path  TEXT,                -- e.g., module path from go.mod (nullable if not available)
  go_version   TEXT,                -- e.g., "go 1.25"
  source_hash  TEXT,                -- content hash of the snapshot for reproducibility
  UNIQUE(project_id, module_path)
);

CREATE TABLE file (
  id            INTEGER PRIMARY KEY,
  container_id  INTEGER NOT NULL REFERENCES container(id) ON DELETE CASCADE,
  rel_path      TEXT NOT NULL,      -- path relative to container root
  pkg_name      TEXT NOT NULL,      -- declared package name
  is_test       INTEGER NOT NULL CHECK (is_test IN (0,1)),
  digest        TEXT,               -- content hash
  size_bytes    INTEGER NOT NULL,
  mod_time      TEXT,               -- capture if available from FS
  UNIQUE(container_id, rel_path)
);

CREATE TABLE package (
  id            INTEGER PRIMARY KEY,
  container_id  INTEGER NOT NULL REFERENCES container(id) ON DELETE CASCADE,
  import_path   TEXT NOT NULL,      -- canonical import path (may include module version info if vendored)
  name          TEXT NOT NULL,      -- package name (may differ from last path elem)
  doc           TEXT,               -- package doc comment (stripped)
  UNIQUE(container_id, import_path)
);

CREATE TABLE pkg_import (
  id           INTEGER PRIMARY KEY,
  package_id   INTEGER NOT NULL REFERENCES package(id) ON DELETE CASCADE,
  path         TEXT NOT NULL,       -- imported path
  alias        TEXT,                -- local alias if present
  is_stdlib    INTEGER NOT NULL DEFAULT 0 CHECK (is_stdlib IN (0,1))
);

CREATE TABLE symbol (
  id            INTEGER PRIMARY KEY,
  package_id    INTEGER NOT NULL REFERENCES package(id) ON DELETE CASCADE,
  file_id       INTEGER REFERENCES file(id) ON DELETE SET NULL,
  kind          TEXT NOT NULL,
  name          TEXT NOT NULL,
  recv_type     TEXT,               -- for methods: canonical receiver type (no pointer/star)
  type_text     TEXT,               -- canonical type/string form for quick display
  doc           TEXT,               -- stripped doc comment at declaration
  span_start    INTEGER,            -- byte offsets in file (optional but useful)
  span_end      INTEGER,
  line          INTEGER,            -- 1-based
  col           INTEGER,            -- 1-based
  exported      INTEGER NOT NULL DEFAULT 0 CHECK (exported IN (0,1))
);

CREATE TABLE signature (
  id            INTEGER PRIMARY KEY,
  symbol_id     INTEGER NOT NULL REFERENCES symbol(id) ON DELETE CASCADE,
  text          TEXT NOT NULL,      -- full signature (pretty-printed)
  params_json   TEXT,               -- JSON array of params (name,type,variadic)
  results_json  TEXT,               -- JSON array of results (name,type)
  type_params_json TEXT             -- JSON array for generics (name,constraints)
);

CREATE TABLE relation (
  id            INTEGER PRIMARY KEY,
  from_symbol_id INTEGER NOT NULL REFERENCES symbol(id) ON DELETE CASCADE,
  to_symbol_id   INTEGER NOT NULL REFERENCES symbol(id) ON DELETE CASCADE,
  kind          TEXT NOT NULL,
  detail        TEXT,               -- optional payload (e.g., method name, position)
  UNIQUE(from_symbol_id, to_symbol_id, kind)
);

CREATE TABLE type_ref (
  id            INTEGER PRIMARY KEY,
  symbol_id     INTEGER NOT NULL REFERENCES symbol(id) ON DELETE CASCADE,
  target_pkg    TEXT,               -- import path if external; NULL if same package
  target_name   TEXT NOT NULL,      -- referenced identifier name
  kind          TEXT NOT NULL,      -- 'type','iface','struct','alias','generic', etc.
  pos_byte      INTEGER             -- optional position for UI
);

CREATE TABLE member (
  id            INTEGER PRIMARY KEY,
  parent_symbol_id INTEGER NOT NULL REFERENCES symbol(id) ON DELETE CASCADE, -- the type
  child_symbol_id  INTEGER REFERENCES symbol(id) ON DELETE SET NULL,         -- the field/method symbol if we modeled it
  name          TEXT NOT NULL,
  exported      INTEGER NOT NULL DEFAULT 0 CHECK (exported IN (0,1)),
  kind          TEXT NOT NULL              -- 'field','method','embedded'
);

CREATE TABLE diagnostic (
  id            INTEGER PRIMARY KEY,
  file_id       INTEGER NOT NULL REFERENCES file(id) ON DELETE CASCADE,
  severity      TEXT NOT NULL,      -- 'info','warn','error'
  code          TEXT,               -- tool/source code (e.g., 'types', 'parser')
  message       TEXT NOT NULL,
  line          INTEGER,
  col           INTEGER
);

CREATE TABLE IF NOT EXISTS language (
  id        TEXT PRIMARY KEY,      -- short, stable key: 'go', 'rust', 'zig', ...
  name      TEXT NOT NULL UNIQUE,  -- display name
  ecosystem TEXT,                  -- optional (e.g., 'JVM', '.NET', 'Node', 'BEAM')
  -- room for future: parser_pack, version, etc.
  CHECK (length(id) > 0)
);

-- Extension → language (fast allowlist)
CREATE TABLE IF NOT EXISTS source_extension (
  ext          TEXT PRIMARY KEY,   -- WITHOUT leading dot, lowercased (e.g., 'go', 'rs')
  language_id  TEXT NOT NULL REFERENCES language(id) ON DELETE RESTRICT,
  is_text      INTEGER NOT NULL DEFAULT 1,      -- 1=text, 0=binary-ish
  is_primary   INTEGER NOT NULL DEFAULT 1,      -- 1=typical source for lang, 0=aux (e.g., headers)
  notes        TEXT
);


-- Basename (no extension) → language
CREATE TABLE IF NOT EXISTS source_basename (
  name         TEXT PRIMARY KEY,   -- exact case-sensitive match (normalize how you like)
  language_id  TEXT NOT NULL REFERENCES language(id) ON DELETE RESTRICT,
  is_text      INTEGER NOT NULL DEFAULT 1,
  notes        TEXT
);

CREATE UNIQUE INDEX ux_symbol_pkg_kind_name_recv
  ON symbol(package_id, kind, name, ifnull(recv_type, ''));


-- Indices for common lookups/hot paths
CREATE INDEX IF NOT EXISTS idx_file_container_path ON file(container_id, rel_path);
CREATE INDEX IF NOT EXISTS idx_pkg_container_import ON package(container_id, import_path);
CREATE INDEX IF NOT EXISTS idx_symbol_pkg_name ON symbol(package_id, name);
CREATE INDEX IF NOT EXISTS idx_symbol_kind ON symbol(kind);
CREATE INDEX IF NOT EXISTS idx_relation_from ON relation(from_symbol_id);
CREATE INDEX IF NOT EXISTS idx_relation_to ON relation(to_symbol_id);
CREATE INDEX IF NOT EXISTS idx_typeref_symbol ON type_ref(symbol_id);
CREATE INDEX IF NOT EXISTS idx_pkg_import_pkg ON pkg_import(package_id);
CREATE INDEX IF NOT EXISTS idx_source_extension_lang ON source_extension(language_id);
CREATE INDEX IF NOT EXISTS idx_source_basename_lang ON source_basename(language_id);

-- Mark DB version
PRAGMA user_version = 1;