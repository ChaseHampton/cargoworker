SELECT name, language_id, is_text, notes
FROM source_basename
WHERE name = ?
LIMIT 1;