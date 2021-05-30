package simplegraph

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLITE                  = "sqlite3"
	WITH_FOREIGN_KEY_PRAGMA = "%s?_foreign_keys=true"
	ID_CONSTRAINT           = "NOT NULL constraint failed: nodes.id"
	UNIQUE_ID_CONSTRAINT    = "UNIQUE constraint failed: nodes.id"
	NO_ROWS_FOUND           = "sql: no rows in result set"
)

type NodeData struct {
	Identifier interface{} `json:"id"`
	Body       interface{}
}

type EdgeData struct {
	Source string
	Target string
	Label  string
}

func resolveDbFileReference(names ...string) (string, error) {
	args := len(names)
	switch args {
	case 1:
		return fmt.Sprintf(WITH_FOREIGN_KEY_PRAGMA, names[0]), nil
	case 2:
		return fmt.Sprintf(WITH_FOREIGN_KEY_PRAGMA, filepath.Join(names[0], names[1])), nil
	default:
		return "", errors.New("invalid database file reference")
	}
}

func evaluate(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Initialize(database ...string) {
	init := func(db *sql.DB) error {
		for _, statement := range strings.Split(Schema, ";") {
			sql := strings.TrimSpace(statement)
			if len(sql) > 0 {
				stmt, err := db.Prepare(sql)
				evaluate(err)
				stmt.Exec()
			}
		}
		return nil
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	init(db)
}

func makeBulkInsertStatement(statement string, inserts int) string {
	pivot := "VALUES"
	parts := strings.Split(strings.TrimSpace(statement), pivot)
	if len(parts) == 2 {
		vals := make([]string, 0, inserts)
		for i := 0; i < inserts; i++ {
			vals = append(vals, parts[1])
		}
		return fmt.Sprintf("%s%s%s", parts[0], pivot, strings.Join(vals, ","))
	}
	return statement
}

func makeBulkEdgeInserts(sources []string, targets []string, properties []string) []interface{} {
	l := len(sources)
	if l != len(targets) && l != len(properties) {
		evaluate(errors.New("unequal edge lists"))
	}
	args := make([]interface{}, 0, l*3)
	for i := 0; i < l; i++ {
		args = append(args, sources[i])
		args = append(args, targets[i])
		args = append(args, properties[i])
	}
	return args
}

func insertMany(nodes []interface{}, database ...string) (int64, error) {
	ins := func(db *sql.DB) (sql.Result, error) {
		stmt, stmtErr := db.Prepare(makeBulkInsertStatement(InsertNode, len(nodes)))
		evaluate(stmtErr)
		return stmt.Exec(nodes...)
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	in, inErr := ins(db)
	if inErr != nil {
		return 0, inErr
	}
	return in.RowsAffected()
}

func insertOne(node string, database ...string) (int64, error) {
	ins := func(db *sql.DB) (sql.Result, error) {
		stmt, stmtErr := db.Prepare(InsertNode)
		evaluate(stmtErr)
		return stmt.Exec(node)
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	in, inErr := ins(db)
	if inErr != nil {
		return 0, inErr
	}
	return in.RowsAffected()
}

func connectMany(edges []interface{}, count int, database ...string) (int64, error) {
	ins := func(db *sql.DB) (sql.Result, error) {
		stmt, stmtErr := db.Prepare(makeBulkInsertStatement(InsertEdge, count))
		evaluate(stmtErr)
		return stmt.Exec(edges...)
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	in, inErr := ins(db)
	if inErr != nil {
		return 0, inErr
	}
	return in.RowsAffected()
}

func needsIdentifier(node []byte) bool {
	var nodeData NodeData
	err := json.Unmarshal(node, &nodeData)
	evaluate(err)
	return nodeData.Identifier == nil
}

func setIdentifier(node []byte, identifier string) []byte {
	closingBraceIdx := bytes.LastIndexByte(node, '}')
	if closingBraceIdx > 0 {
		addId := []byte(fmt.Sprintf(", \"id\": %q", identifier))
		node = append(node[:closingBraceIdx], addId...)
		node = append(node, '}')
	}
	return node
}

func AddNode(identifier string, node []byte, database ...string) (int64, error) {
	if needsIdentifier(node) {
		return insertOne(string(setIdentifier(node, identifier)), database...)
	}
	return insertOne(string(node), database...)
}

func AddNodes(identifiers []string, nodes [][]byte, database ...string) (int64, error) {
	l := len(nodes)
	if l != len(identifiers) {
		evaluate(errors.New("unequal node, identifier lists"))
	}
	args := make([]interface{}, l)
	for i := 0; i < l; i++ {
		if needsIdentifier(nodes[i]) {
			args[i] = string(setIdentifier(nodes[i], identifiers[i]))
		} else {
			args[i] = string(nodes[i])
		}

	}
	return insertMany(args, database...)
}

func ConnectNodesWithProperties(sourceId string, targetId string, properties []byte, database ...string) (int64, error) {
	connect := func(db *sql.DB) (sql.Result, error) {
		stmt, stmtErr := db.Prepare(InsertEdge)
		evaluate(stmtErr)
		return stmt.Exec(sourceId, targetId, string(properties))
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	cx, cxErr := connect(db)
	if cxErr != nil {
		return 0, cxErr
	}
	return cx.RowsAffected()
}

func ConnectNodes(sourceId string, targetId string, database ...string) (int64, error) {
	return ConnectNodesWithProperties(sourceId, targetId, []byte(`{}`), database...)
}

func BulkConnectNodesWithProperties(sources []string, targets []string, properties []string, database ...string) (int64, error) {
	l := len(sources)
	if l != len(targets) && l != len(properties) {
		evaluate(errors.New("unequal source, target, properties lists"))
	}
	return connectMany(makeBulkEdgeInserts(sources, targets, properties), l, database...)
}

func BulkConnectNodes(sources []string, targets []string, database ...string) (int64, error) {
	l := len(sources)
	props := make([]string, 0, l)
	for i := 0; i < l; i++ {
		props = append(props, `{}`)
	}
	return BulkConnectNodesWithProperties(sources, targets, props, database...)
}

func RemoveNodes(identifiers []string, database ...string) bool {
	delete := func(db *sql.DB) bool {
		edgeStmt, edgeErr := db.Prepare(DeleteEdge)
		evaluate(edgeErr)
		nodeStmt, nodeErr := db.Prepare(DeleteNode)
		evaluate(nodeErr)
		tx, txErr := db.Begin()
		evaluate(txErr)

		var err error
		for _, identifier := range identifiers {
			_, err = tx.Stmt(edgeStmt).Exec(identifier, identifier)
			if err != nil {
				tx.Rollback()
				return false
			}
			_, err = tx.Stmt(nodeStmt).Exec(identifier)
			if err != nil {
				tx.Rollback()
				return false
			}
		}
		tx.Commit()
		return true
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	return delete(db)
}

func FindNode(identifier string, database ...string) (string, error) {
	find := func(db *sql.DB) (string, error) {
		stmt, err := db.Prepare(SearchNodeById)
		evaluate(err)
		defer stmt.Close()
		var body string
		err = stmt.QueryRow(identifier).Scan(&body)
		if err == sql.ErrNoRows {
			return "", err
		}
		evaluate(err)
		return body, nil
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	return find(db)
}

func UpdateNodeBody(identifier string, body string, database ...string) error {
	update := func(db *sql.DB) error {
		stmt, err := db.Prepare(UpdateNode)
		evaluate(err)
		defer stmt.Close()
		_, err = stmt.Exec(body, identifier)
		return err
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	return update(db)
}

func UpsertNode(identifier string, body string, database ...string) error {
	update := []byte(body)
	node, err := FindNode(identifier, database...)
	if node == "" && err == sql.ErrNoRows {
		_, err = AddNode(identifier, update, database...)
		return err
	} else {
		if needsIdentifier(update) {
			return UpdateNodeBody(identifier, string(setIdentifier(update, identifier)), database...)
		}
		return UpdateNodeBody(identifier, body, database...)
	}
}

func generateWhereClauseForSearch(properties map[string]string, predicate string) string {
	clauses := []string{}
	for key := range properties {
		clause := strings.Builder{}
		fmt.Fprintf(&clause, "json_extract(body, '$.%s') %s ?", key, predicate)
		clauses = append(clauses, clause.String())
	}
	return strings.Join(clauses, " AND ")
}

func generateSearchEquals(properties map[string]string) string {
	return generateWhereClauseForSearch(properties, "=")
}

func generateSearchLike(properties map[string]string) string {
	return generateWhereClauseForSearch(properties, "LIKE")
}

func generateSearchStatement(properties map[string]string, equality bool) string {
	var where string
	if equality {
		where = generateSearchEquals(properties)
	} else {
		where = generateSearchLike(properties)
	}
	return fmt.Sprintf("%s %s", strings.TrimSpace(SearchNode), where)
}

func generateSearchBindings(properties map[string]string, startsWith bool, contains bool) []string {
	bindings := []string{}
	for _, val := range properties {
		var binding string
		if startsWith {
			binding = fmt.Sprintf("%s%%", val)
		} else if contains {
			binding = fmt.Sprintf("%%%s%%", val)
		} else {
			binding = val
		}
		bindings = append(bindings, binding)
	}
	return bindings
}

func convertSearchBindingsToParameters(bindings []string) []interface{} {
	params := make([]interface{}, len(bindings))
	for i, binding := range bindings {
		params[i] = binding
	}
	return params
}

func FindNodes(properties map[string]string, startsWith bool, contains bool, database ...string) ([]string, error) {
	var statement string
	if startsWith || contains {
		statement = generateSearchStatement(properties, false)
	} else {
		statement = generateSearchStatement(properties, true)
	}
	bindings := generateSearchBindings(properties, startsWith, contains)

	find := func(db *sql.DB) ([]string, error) {
		stmt, stmtErr := db.Prepare(statement)
		evaluate(stmtErr)
		defer stmt.Close()

		results := []string{}
		rows, err := stmt.Query(convertSearchBindingsToParameters(bindings)...)
		if err != nil {
			results = append(results, "")
			return results, err
		}
		defer rows.Close()
		for rows.Next() {
			var body string
			err = rows.Scan(&body)
			if err != nil {
				results = append(results, "")
				return results, err
			}
			results = append(results, body)
		}
		err = rows.Err()
		return results, err
	}

	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	return find(db)
}

func traverse(source string, statement string, target string) func(*sql.DB) ([]string, error) {
	return func(db *sql.DB) ([]string, error) {
		stmt, stmtErr := db.Prepare(statement)
		evaluate(stmtErr)
		defer stmt.Close()

		results := []string{}
		rows, err := stmt.Query(source)
		if err != nil {
			results = append(results, "")
			return results, err
		}
		defer rows.Close()
		for rows.Next() {
			var identifier string
			err = rows.Scan(&identifier)
			if err != nil {
				results = append(results, "")
				return results, err
			}
			results = append(results, identifier)
			if len(target) > 0 && identifier == target {
				break
			}
		}
		err = rows.Err()
		return results, err
	}
}

func TraverseFromTo(source string, target string, traversal string, database ...string) ([]string, error) {
	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	fn := traverse(source, traversal, target)
	return fn(db)
}

func TraverseFrom(source string, traversal string, database ...string) ([]string, error) {
	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	fn := traverse(source, traversal, "")
	return fn(db)
}

func neighbors(statement string, queryBinding func(*sql.Stmt) (*sql.Rows, error)) func(*sql.DB) ([]EdgeData, error) {
	return func(db *sql.DB) ([]EdgeData, error) {
		stmt, stmtErr := db.Prepare(statement)
		evaluate(stmtErr)
		defer stmt.Close()

		results := []EdgeData{}
		rows, err := queryBinding(stmt)
		if err != nil {
			results = append(results, EdgeData{})
			return results, err
		}
		defer rows.Close()
		for rows.Next() {
			var result EdgeData
			var source string
			var target string
			var label string
			err = rows.Scan(&source, &target, &label)
			if err != nil {
				results = append(results, result)
				return results, err
			}
			result.Source = source
			result.Target = target
			if len(label) > 0 {
				result.Label = label
			}
			results = append(results, result)
		}
		err = rows.Err()
		return results, err
	}
}

func getConnectionsOneWay(identifier string, direction string, database ...string) ([]EdgeData, error) {
	query := func(stmt *sql.Stmt) (*sql.Rows, error) {
		return stmt.Query(identifier)
	}
	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	fn := neighbors(direction, query)
	return fn(db)
}

func ConnectionsIn(identifier string, database ...string) ([]EdgeData, error) {
	return getConnectionsOneWay(identifier, SearchEdgesInbound, database...)
}

func ConnectionsOut(identifier string, database ...string) ([]EdgeData, error) {
	return getConnectionsOneWay(identifier, SearchEdgesOutbound, database...)
}

func Connections(identifier string, database ...string) ([]EdgeData, error) {
	query := func(stmt *sql.Stmt) (*sql.Rows, error) {
		return stmt.Query(identifier, identifier)
	}
	dbReference, err := resolveDbFileReference(database...)
	evaluate(err)
	db, dbErr := sql.Open(SQLITE, dbReference)
	evaluate(dbErr)
	defer db.Close()
	fn := neighbors(SearchEdges, query)
	return fn(db)
}
