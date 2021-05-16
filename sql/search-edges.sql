SELECT * FROM edges WHERE source = ? 
UNION
SELECT * FROM edges WHERE target = ?