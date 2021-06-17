WITH RECURSIVE traverse(x, y, obj) AS (
  SELECT ?, '()', '{}'
  UNION
  SELECT id, '()', body FROM nodes JOIN traverse ON id = x
  UNION
  SELECT target, '->', properties FROM edges JOIN traverse ON source = x
) SELECT x, y, obj FROM traverse;
