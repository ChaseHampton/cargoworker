-- Languages
INSERT OR IGNORE INTO language (id, name, ecosystem) VALUES
  ('go','Go',''),
  ('c','C',''),
  ('cpp','C++',''),
  ('objc','Objective-C','Apple'),
  ('objcxx','Objective-C++','Apple'),
  ('rust','Rust',''),
  ('zig','Zig',''),
  ('gleam','Gleam','BEAM'),
  ('erlang','Erlang','BEAM'),
  ('elixir','Elixir','BEAM'),
  ('haskell','Haskell',''),
  ('ocaml','OCaml',''),
  ('fsharp','F#','.NET'),
  ('csharp','C#','.NET'),
  ('java','Java','JVM'),
  ('kotlin','Kotlin','JVM'),
  ('scala','Scala','JVM'),
  ('groovy','Groovy','JVM'),
  ('python','Python',''),
  ('ruby','Ruby',''),
  ('php','PHP',''),
  ('js','JavaScript','Node'),
  ('ts','TypeScript','Node'),
  ('swift','Swift',''),
  ('dart','Dart',''),
  ('r','R',''),
  ('lua','Lua',''),
  ('shell','Shell',''),
  ('powershell','PowerShell','.NET'),
  ('fish','Fish',''),
  ('sql','SQL',''),
  ('proto','Protocol Buffers',''),
  ('thrift','Thrift',''),
  ('graphql','GraphQL',''),
  ('nim','Nim',''),
  ('reason','ReasonML',''),
  ('ziggy','Zig (alt)',''); -- placeholder if you ever want aliases (optional)

-- Extensions (primary sources first)
INSERT OR IGNORE INTO source_extension (ext, language_id, is_text, is_primary, notes) VALUES
  -- Go
  ('go','go',1,1,'Go source'),
  ('mod','go',1,0,'Go module file'),
  ('sum','go',1,0,'Go module lock'),
  -- C / C++
  ('c','c',1,1,'C source'),
  ('h','c',1,0,'C/C++ header'),
  ('cc','cpp',1,1,'C++ source'),
  ('cpp','cpp',1,1,'C++ source'),
  ('cxx','cpp',1,1,'C++ source'),
  ('hpp','cpp',1,0,'C++ header'),
  ('hh','cpp',1,0,'C++ header'),
  -- Objective-C / Objective-C++
  ('m','objc',1,1,'Objective-C source'),
  ('mm','objcxx',1,1,'Objective-C++ source'),
  -- Rust
  ('rs','rust',1,1,'Rust source'),
  -- Zig / Gleam
  ('zig','zig',1,1,'Zig source'),
  ('gleam','gleam',1,1,'Gleam source'),
  -- BEAM (Erlang/Elixir)
  ('erl','erlang',1,1,'Erlang source'),
  ('hrl','erlang',1,0,'Erlang header'),
  ('ex','elixir',1,1,'Elixir source'),
  ('exs','elixir',1,1,'Elixir script/test'),
  -- Haskell / OCaml / F#
  ('hs','haskell',1,1,'Haskell source'),
  ('lhs','haskell',1,0,'Literate Haskell'),
  ('ml','ocaml',1,1,'OCaml impl'),
  ('mli','ocaml',1,0,'OCaml interface'),
  ('fs','fsharp',1,1,'F# source'),
  ('fsi','fsharp',1,0,'F# signature'),
  -- .NET
  ('cs','csharp',1,1,'C# source'),
  -- JVM family
  ('java','java',1,1,'Java source'),
  ('kt','kotlin',1,1,'Kotlin source'),
  ('kts','kotlin',1,1,'Kotlin script'),
  ('scala','scala',1,1,'Scala source'),
  ('sc','scala',1,0,'Scala script'),
  ('groovy','groovy',1,1,'Groovy source'),
  -- Python / Ruby / PHP / Lua / R
  ('py','python',1,1,'Python source'),
  ('pyi','python',1,0,'Python type stubs'),
  ('rb','ruby',1,1,'Ruby source'),
  ('php','php',1,1,'PHP source'),
  ('lua','lua',1,1,'Lua source'),
  ('r','r',1,1,'R source'),
  ('R','r',1,1,'R source (capital R)'),
  -- Web/TS/JS
  ('js','js',1,1,'JavaScript'),
  ('mjs','js',1,1,'ES module JS'),
  ('cjs','js',1,1,'CommonJS'),
  ('jsx','js',1,1,'React JSX (JS)'),
  ('ts','ts',1,1,'TypeScript'),
  ('tsx','ts',1,1,'React TSX'),
  -- Swift / Dart
  ('swift','swift',1,1,'Swift'),
  ('dart','dart',1,1,'Dart'),
  -- SQL / data DSLs
  ('sql','sql',1,1,'SQL'),
  ('prisma','sql',1,0,'Prisma schema'),
  -- IDLs
  ('proto','proto',1,1,'Protocol Buffers'),
  ('thrift','thrift',1,1,'Thrift'),
  ('graphql','graphql',1,1,'GraphQL SDL'),
  ('gql','graphql',1,1,'GraphQL SDL'),
  -- Shells
  ('sh','shell',1,1,'POSIX shell'),
  ('bash','shell',1,1,'Bash'),
  ('zsh','shell',1,1,'Zsh'),
  ('fish','fish',1,1,'Fish'),
  ('ksh','shell',1,1,'KornShell'),
  ('ps1','powershell',1,1,'PowerShell'),
  -- Misc often-text sources you might want to treat as code-ish
  ('make','shell',1,0,'Make includes, rarely used'),
  ('cmake','cpp',1,0,'CMake scripts'),
  ('gradle','groovy',1,0,'Gradle build script'),
  ('sbt','scala',1,0,'SBT build'),
  ('nim','nim',1,1,'Nim source');

-- Basename (no extension)
INSERT OR IGNORE INTO source_basename (name, language_id, is_text, notes) VALUES
  ('Makefile','shell',1,'Make build file'),
  ('GNUmakefile','shell',1,'GNU Make build file'),
  ('CMakeLists.txt','cpp',1,'CMake project file'),
  ('Dockerfile','shell',1,'Docker build file'),
  ('BUILD','python',1,'Bazel BUILD (Starlark)'),
  ('BUILD.bazel','python',1,'Bazel BUILD (Starlark)'),
  ('WORKSPACE','python',1,'Bazel WORKSPACE (Starlark)'),
  ('Justfile','shell',1,'just taskfile'),
  ('Rakefile','ruby',1,'Ruby rake'),
  ('Gemfile','ruby',1,'Ruby bundler'),
  ('Pipfile','python',1,'Pipenv'),
  ('requirements.txt','python',1,'Python requirements'),
  ('tsconfig.json','ts',1,'TypeScript config'),
  ('package.json','js',1,'Node metadata');

PRAGMA user_version = 3;