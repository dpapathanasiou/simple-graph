WITH RECURSIVE traverse(id) AS (
  SELECT :source
  UNION
  SELECT target FROM edges JOIN traverse ON source = id
) SELECT id FROM traverse;
