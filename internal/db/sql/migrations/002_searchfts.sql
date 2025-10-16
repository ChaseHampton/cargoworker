PRAGMA foreign_keys = ON;

CREATE VIRTUAL TABLE search_fts USING fts5(
  name,
  pkg,
  kind,
  doc,
  content='symbol',
  content_rowid='id',
  tokenize='unicode61'
);

INSERT INTO search_fts(rowid, name, pkg, kind, doc)
SELECT s.id,
       s.name,
       (SELECT import_path FROM package p WHERE p.id = s.package_id),
       s.kind,
       COALESCE(s.doc,'')
FROM symbol s;

CREATE TRIGGER symbol_ai AFTER INSERT ON symbol BEGIN
  INSERT INTO search_fts(rowid, name, pkg, kind, doc)
  VALUES (
    new.id,
    new.name,
    (SELECT import_path FROM package p WHERE p.id = new.package_id),
    new.kind,
    COALESCE(new.doc,'')
  );
END;

CREATE TRIGGER symbol_ad AFTER DELETE ON symbol BEGIN
  INSERT INTO search_fts(search_fts, rowid, name, pkg, kind, doc)
  VALUES ('delete', old.id, old.name,
          (SELECT import_path FROM package p WHERE p.id = old.package_id),
          old.kind, COALESCE(old.doc,''));
END;

CREATE TRIGGER symbol_au AFTER UPDATE ON symbol BEGIN
  INSERT INTO search_fts(search_fts, rowid, name, pkg, kind, doc)
  VALUES ('delete', old.id, old.name,
          (SELECT import_path FROM package p WHERE p.id = old.package_id),
          old.kind, COALESCE(old.doc,''));
  INSERT INTO search_fts(rowid, name, pkg, kind, doc)
  VALUES (
    new.id,
    new.name,
    (SELECT import_path FROM package p WHERE p.id = new.package_id),
    new.kind,
    COALESCE(new.doc,'')
  );
END;

PRAGMA user_version = 2;
