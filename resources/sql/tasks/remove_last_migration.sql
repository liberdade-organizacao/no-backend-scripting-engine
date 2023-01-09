DELETE FROM migrations 
WHERE id=(
    SELECT id FROM migrations ORDER BY created_at DESC LIMIT 1
) 
RETURNING *;
