package simplegraph

const (
    DeleteEdge = `DELETE FROM edges WHERE source = ? OR target = ?
`

    DeleteNode = `DELETE FROM nodes WHERE id = ?
`

    InsertEdge = `INSERT INTO edges VALUES(?, ?, json(?))
`

    InsertNode = `INSERT INTO nodes VALUES(json(?))
`

    Schema = `CREATE TABLE IF NOT EXISTS nodes (
    body TEXT,
    id   TEXT GENERATED ALWAYS AS (json_extract(body, '$.id')) VIRTUAL NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS id_idx ON nodes(id);

CREATE TABLE IF NOT EXISTS edges (
    source     TEXT,
    target     TEXT,
    properties TEXT,
    FOREIGN KEY(source) REFERENCES nodes(id),
    FOREIGN KEY(target) REFERENCES nodes(id)
);

CREATE INDEX IF NOT EXISTS source_idx ON edges(source);
CREATE INDEX IF NOT EXISTS target_idx ON edges(target);
`

    SearchEdgesInbound = `SELECT * FROM edges WHERE source = ?
`

    SearchEdgesOutbound = `SELECT * FROM edges WHERE target = ?
`

    SearchEdges = `SELECT * FROM edges WHERE source = ? 
UNION
SELECT * FROM edges WHERE target = ?
`

    SearchNodeById = `SELECT body FROM nodes WHERE json_extract(body, '$.id') = ?
`

    SearchNode = `SELECT body FROM nodes WHERE 
`

    TraverseInbound = `WITH RECURSIVE traverse(id) AS (
  SELECT ?
  UNION
  SELECT source FROM edges JOIN traverse ON target = id
) SELECT id FROM traverse;
`

    TraverseOutbound = `WITH RECURSIVE traverse(id) AS (
  SELECT ?
  UNION
  SELECT target FROM edges JOIN traverse ON source = id
) SELECT id FROM traverse;
`

    Traverse = `WITH RECURSIVE traverse(id) AS (
  SELECT ?
  UNION
  SELECT source FROM edges JOIN traverse ON target = id
  UNION
  SELECT target FROM edges JOIN traverse ON source = id
) SELECT id FROM traverse;
`

    UpdateNode = `UPDATE nodes SET body = json(?) WHERE id = ?
`

)
