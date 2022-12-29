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
    UNIQUE(source, target, properties) ON CONFLICT REPLACE,
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

	UpdateNode = `UPDATE nodes SET body = json(?) WHERE id = ?
`

	SearchNodeTemplate = `SELECT {{ .ResultColumn }} -- id|body
From nodes{{ if .Tree }}, json_tree(body{{ if .Key }}, '$.{{ .Key }}'{{ end }}){{ end }}{{ if .SearchClauses }}
WHERE {{ range .SearchClauses }}
    {{ . }}
{{ end }}{{ end }}
`

	SearchWhereTemplate = `{{ if .AndOr }}{{ .AndOr }}{{ end }}
{{ if .IdLookup }}id = ?{{ end }}
{{ if .KeyValue }}json_extract(body, '$.{{ .Key }}') {{ .Predicate }} ?{{ end }}
{{ if .Tree }}{{ if .Key }}(json_tree.key='{{ .Key }}' and {{ end }}json_tree.value {{ .Predicate }} ?{{ if .Key }}){{ end }}{{ end }}
`

	TraverseTemplate = `WITH RECURSIVE traverse(x{{ if .WithBodies }}, y, obj{{ end }}) AS (
  SELECT id{{ if .WithBodies }}, '()', body {{ end }} FROM nodes WHERE id = ?
  UNION
  SELECT id{{ if .WithBodies }}, '()', body {{ end }} FROM nodes JOIN traverse ON id = x
  {{ if .Inbound }}UNION
  SELECT source{{ if .WithBodies }}, '<-', properties {{ end }} from edges join traverse on target = x{{ end }}
  {{ if .Outbound }}UNION
  SELECT target{{ if .WithBodies }}, '->', properties {{ end }} from edges join traverse on source = x{{ end }}
) SELECT x{{ if .WithBodies }}, y, obj {{ end }} FROM traverse;
`
)
