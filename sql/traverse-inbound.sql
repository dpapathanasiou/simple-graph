WITH RECURSIVE traverse(id) AS (
  SELECT ?
  UNION
  SELECT source FROM edges JOIN traverse ON target = id
) SELECT id FROM traverse;
