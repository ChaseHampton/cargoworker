SELECT ext, language_id, is_text, is_primary, notes
FROM source_extension
WHERE ext = ?
LIMIT 1;