package simplegraph

import (
	"os"
	"testing"
)

const (
	apple    = `{"name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976","id":"1"}`
	woz      = `{"id":"2","name":"Steve Wozniak","type":["person","engineer","founder"]}`
	wozNick  = `{"name":"Steve Wozniak","type":["person","engineer","founder"],"nickname":"Woz","id":"2"}`
	jobs     = `{"id":"3","name":"Steve Jobs","type":["person","designer","founder"]}`
	wayne    = `{"id":"4","name":"Ronald Wayne","type":["person","administrator","founder"]}`
	markkula = `{"name":"Mike Markkula","type":["person","investor"]}`
	founded  = `{"action":"founded"}`
	invested = `{"action":"invested","equity":80000,"debt":170000}`
	divested = `{"action":"divested","amount":800,"date":"April 12, 1976"}`
)

func ErrorMatches(actual error, expected string) bool {
	if actual == nil {
		return expected == ""
	}
	if expected == "" {
		return false
	}
	return actual.Error() == expected
}

func TestResolveDbFileReference(t *testing.T) {
	path := "/tmp/database.sqlite?_foreign_keys=true"
	actualPath, actualPathErr := resolveDbFileReference("/tmp", "database.sqlite")
	if actualPath != path {
		t.Errorf("resolveDbFileReference(\"/tmp\", \"database.sqlite\") = %q but expected %q", actualPath, path)
	}
	if actualPathErr != nil {
		t.Errorf("resolveDbFileReference(\"/tmp\", \"database.sqlite\") = %q but expected nil", actualPathErr.Error())
	}

	file := "database.sqlite?_foreign_keys=true"
	actualFile, actualFileErr := resolveDbFileReference("database.sqlite")
	if actualFile != file {
		t.Errorf("resolveDbFileReference(\"database.sqlite\") = %q but expected %q", actualFile, file)
	}
	if actualFileErr != nil {
		t.Errorf("resolveDbFileReference(\"database.sqlite\") = %q but expected nil", actualFileErr.Error())
	}

	empty := "invalid database file reference"
	emptyFile, emptyFileErr := resolveDbFileReference()
	if emptyFile != "" {
		t.Errorf("resolveDbFileReference() = %q but expected %q", emptyFile, "")
	}
	if !ErrorMatches(emptyFileErr, empty) {
		t.Errorf("resolveDbFileReference() = %q but expected %q", emptyFileErr.Error(), empty)
	}
}

func arrayContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func resultContains(slice []GraphData, val GraphData) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func TestGenerateSearchStatement(t *testing.T) {
	kvNameLikeClause := "\n\njson_extract(body, '$.name') LIKE ?\n\n"
	kvNameLike := GenerateWhereClause(&WhereClause{KeyValue: true, Key: "name", Predicate: "LIKE"})
	if kvNameLike != kvNameLikeClause {
		t.Errorf("generateWhereClause() = %q but expected %q", kvNameLike, kvNameLikeClause)
	}

	treeValueWithOrClause := "OR\n\n\njson_tree.value = ?\n"
	treeValueWithOr := GenerateWhereClause(&WhereClause{AndOr: "OR", Tree: true, Predicate: "="})
	if treeValueWithOr != treeValueWithOrClause {
		t.Errorf("generateWhereClause() = %q but expected %q", treeValueWithOr, treeValueWithOrClause)
	}
}

func TestMakeBulkInsertStatement(t *testing.T) {
	expected := `INSERT INTO nodes VALUES(json(?))`
	actual := makeBulkInsertStatement(InsertNode, 1)
	if expected != actual {
		t.Errorf("generateBulkInsertStatement() = %q but expected %q", actual, expected)
	}

	expected = `INSERT INTO nodes VALUES(json(?)),(json(?)),(json(?))`
	actual = makeBulkInsertStatement(InsertNode, 3)
	if expected != actual {
		t.Errorf("generateBulkInsertStatement() = %q but expected %q", actual, expected)
	}

	expected = `INSERT INTO edges VALUES(?, ?, json(?)),(?, ?, json(?))`
	actual = makeBulkInsertStatement(InsertEdge, 2)
	if expected != actual {
		t.Errorf("generateBulkInsertStatement() = %q but expected %q", actual, expected)
	}
}

func TestMakeBulkEdgeInserts(t *testing.T) {
	expected := []string{"3", "1", founded, "4", "1", `{}`}
	for i, actual := range makeBulkEdgeInserts([]string{"3", "4"}, []string{"1", "1"}, []string{founded, `{}`}) {
		if expected[i] != actual {
			t.Errorf("generateBulkInsertStatement() = %q but expected %q", actual, expected[i])
		}
	}
}

func TestNodeDataInspection(t *testing.T) {
	missing := needsIdentifier([]byte(`{"status": 404,"result": "error", "reason": "Not found"}`))
	if !missing {
		t.Errorf("needsIdentifier() said false but expected true")
	}

	alsoMissing := needsIdentifier([]byte(`{"status": 404,"result": "error", "logger": {"id": "9c26f784-b0d6-45ed-aba4-7c333f78babf"}, "reason": "Not found"}`))
	if !alsoMissing {
		t.Errorf("needsIdentifier() said false but expected true")
	}

	present := needsIdentifier([]byte(`{"status": 404,"result": "error", "id": "16fd2706-8baf-433b-82eb-8c7fada847da", "reason": "Not found"}`))
	if present {
		t.Errorf("needsIdentifier() said true but expected false")
	}
}

func TestInitializeAndCrudAndSearch(t *testing.T) {
	file := "testdb.sqlite3"
	Initialize(file)
	defer os.Remove(file)

	fs, fsErr := os.Lstat(file)
	if fs.Name() != file {
		t.Errorf("Initialize() produced %q but expected %q", fs.Name(), file)
	}
	if fsErr != nil {
		t.Errorf("Initialize() produced error %q but expected nil", fsErr.Error())
	}

	count, err := AddNode("1", []byte(apple), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("2", []byte(woz), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("3", []byte(jobs), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("4", []byte(wayne), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("5", []byte(markkula), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("1", []byte(apple), file)
	if count != 0 && !ErrorMatches(err, UNIQUE_ID_CONSTRAINT) {
		t.Errorf("AddNode() inserted %d,%q but expected 0,%q", count, err.Error(), UNIQUE_ID_CONSTRAINT)
	}

	count, err = AddNode("2", []byte(woz), file)
	if count != 0 && !ErrorMatches(err, ID_CONSTRAINT) {
		t.Errorf("AddNode() inserted %d,%q but expected 0,%q", count, err.Error(), ID_CONSTRAINT)
	}

	count, err = ConnectNodesWithProperties("2", "1", []byte(founded), file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodesWithProperties("3", "1", []byte(founded), file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodesWithProperties("4", "1", []byte(founded), file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodesWithProperties("5", "1", []byte(invested), file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodesWithProperties("1", "4", []byte(divested), file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodes("2", "3", file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodes() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	node, err := FindNode("1", file)
	if node != apple && err != nil {
		t.Errorf("FindNode() produced %q,%q but expected %q,nil", node, err.Error(), apple)
	}

	node, err = FindNode("7", file)
	if node != "" && !ErrorMatches(err, NO_ROWS_FOUND) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", NO_ROWS_FOUND)
	}

	kvNameLike := GenerateWhereClause(&WhereClause{KeyValue: true, Key: "name", Predicate: "LIKE"})
	statement := GenerateSearchStatement(&SearchQuery{ResultColumn: "body", SearchClauses: []string{kvNameLike}})

	nodes, err := FindNodes(statement, []string{"Steve%"}, file)
	if err != nil {
		t.Errorf("FindNodes() produced an error %s but expected nil", err.Error())
	}
	if !arrayContains(nodes, woz) {
		t.Errorf("FindNodes() did not return %s as expected", woz)
	}
	if !arrayContains(nodes, jobs) {
		t.Errorf("FindNodes() did not return %s as expected", jobs)
	}

	err = UpdateNodeBody("2", wozNick, file)
	if err != nil {
		t.Errorf("UpdateNodeBody() produced %q but expected nil", err.Error())
	}

	err = UpsertNode("1", apple, file)
	if err != nil {
		t.Errorf("UpsertNode() produced %q but expected nil", err.Error())
	}

	node, err = FindNode("2", file)
	if node != wozNick && err != nil {
		t.Errorf("FindNode() produced %q,%q but expected %q,nil", node, err.Error(), apple)
	}

	arrayType := GenerateWhereClause(&WhereClause{Tree: true, Predicate: "="})
	statement = GenerateSearchStatement(&SearchQuery{ResultColumn: "body", Tree: true, Key: "type", SearchClauses: []string{arrayType}})

	nodes, err = FindNodes(statement, []string{"founder"}, file)
	if err != nil {
		t.Errorf("FindNodes() produced an error %s but expected nil", err.Error())
	}
	if !arrayContains(nodes, wozNick) {
		t.Errorf("FindNodes() did not return %s as expected", woz)
	}
	if !arrayContains(nodes, jobs) {
		t.Errorf("FindNodes() did not return %s as expected", jobs)
	}
	if !arrayContains(nodes, wayne) {
		t.Errorf("FindNodes() did not return %s as expected", wayne)
	}

	basicTraversal := GenerateTraversal(&Traversal{WithBodies: false, Inbound: true, Outbound: true})
	idList, traverseErr := TraverseFromTo("2", "3", basicTraversal, file)
	if traverseErr != nil {
		t.Errorf("TraverseFromTo() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"2", "1", "3"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFromTo() did not return %s as expected", expectedId)
		}
	}

	basicTraversalInbound := GenerateTraversal(&Traversal{WithBodies: false, Inbound: true, Outbound: false})
	idList, traverseErr = TraverseFrom("5", basicTraversalInbound, file)
	if traverseErr != nil {
		t.Errorf("TraverseFrom() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"5"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFrom() did not return %s as expected", expectedId)
		}
	}

	basicTraversalOutbound := GenerateTraversal(&Traversal{WithBodies: false, Inbound: false, Outbound: true})
	idList, traverseErr = TraverseFrom("5", basicTraversalOutbound, file)
	if traverseErr != nil {
		t.Errorf("TraverseFrom() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"5", "1", "4"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFrom() did not return %s as expected", expectedId)
		}
	}

	idList, traverseErr = TraverseFrom("5", basicTraversal, file)
	if traverseErr != nil {
		t.Errorf("TraverseFrom() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"5", "1", "2", "3", "4"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFrom() did not return %s as expected", expectedId)
		}
	}

	basicTraversalWithBodies := GenerateTraversal(&Traversal{WithBodies: true, Inbound: true, Outbound: true})
	nilNode := NodeData{Identifier: nil, Body: nil}
	bodies, traverseWithErr := TraverseWithBodiesFromTo("2", "3", basicTraversalWithBodies, file)
	if traverseWithErr != nil {
		t.Errorf("TraverseWithBodiesFromTo() produced an error %s but expected nil", traverseWithErr.Error())
	}
	for _, expectedObject := range []GraphData{
		{Node: NodeData{Identifier: "2", Body: wozNick}},
		{Node: nilNode, Edge: EdgeData{Source: "2", Target: "1", Label: founded}},
		{Node: nilNode, Edge: EdgeData{Source: "2", Target: "3", Label: "{}"}},
		{Node: NodeData{Identifier: "1", Body: apple}},
		{Node: nilNode, Edge: EdgeData{Source: "2", Target: "1", Label: founded}},
		{Node: nilNode, Edge: EdgeData{Source: "3", Target: "1", Label: founded}},
		{Node: nilNode, Edge: EdgeData{Source: "4", Target: "1", Label: founded}},
		{Node: nilNode, Edge: EdgeData{Source: "5", Target: "1", Label: invested}},
		{Node: nilNode, Edge: EdgeData{Source: "1", Target: "4", Label: divested}},
		{Node: NodeData{Identifier: "3", Body: jobs}}} {
		if !resultContains(bodies, expectedObject) {
			t.Errorf("TraverseWithBodiesFromTo() did not return %v as expected", expectedObject)
		}
	}

	edges, err := ConnectionsIn("1", file)
	if err != nil {
		t.Errorf("ConnectionsIn() produced an error %s but expected nil", err.Error())
	}
	expected := []EdgeData{{"1", "4", divested}}
	for i, exp := range expected {
		if edges[i] != exp {
			t.Errorf("ConnectionsIn() produced %q but expected %q", edges[i], exp)
		}
	}

	edges, err = ConnectionsOut("1", file)
	if err != nil {
		t.Errorf("ConnectionsIn() produced an error %s but expected nil", err.Error())
	}
	expected = []EdgeData{{"2", "1", founded},
		{"3", "1", founded},
		{"4", "1", founded},
		{"5", "1", invested}}
	for i, exp := range expected {
		if edges[i] != exp {
			t.Errorf("ConnectionsOut() produced %q but expected %q", edges[i], exp)
		}
	}

	edges, err = Connections("1", file)
	if err != nil {
		t.Errorf("Connections() produced an error %s but expected nil", err.Error())
	}
	expected = []EdgeData{{"1", "4", divested},
		{"2", "1", founded},
		{"3", "1", founded},
		{"4", "1", founded},
		{"5", "1", invested}}
	for i, exp := range expected {
		if edges[i] != exp {
			t.Errorf("ConnectionsOut() produced %q but expected %q", edges[i], exp)
		}
	}

	if !RemoveNodes([]string{"2", "4"}, file) {
		t.Error("RemoveNodes() returned false but expected true")
	}

	node, err = FindNode("2", file)
	if node != "" && !ErrorMatches(err, NO_ROWS_FOUND) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", NO_ROWS_FOUND)
	}

	node, err = FindNode("4", file)
	if node != "" && !ErrorMatches(err, NO_ROWS_FOUND) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", NO_ROWS_FOUND)
	}
}

func TestBulkInserts(t *testing.T) {
	file := "testdb.sqlite3"
	Initialize(file)
	defer os.Remove(file)

	fs, fsErr := os.Lstat(file)
	if fs.Name() != file {
		t.Errorf("Initialize() produced %q but expected %q", fs.Name(), file)
	}
	if fsErr != nil {
		t.Errorf("Initialize() produced error %q but expected nil", fsErr.Error())
	}

	nodes := [][]byte{[]byte(apple),
		[]byte(woz),
		[]byte(jobs),
		[]byte(wayne),
		[]byte(markkula)}
	ids := []string{"1", "2", "3", "4", "5"}

	count, err := AddNodes(ids, nodes, file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	sources := []string{"2", "3", "4", "5", "1"}
	targets := []string{"1", "1", "1", "1", "4"}
	properties := []string{founded,
		founded,
		founded,
		invested,
		divested}

	count, err = BulkConnectNodesWithProperties(sources, targets, properties, file)
	if count != 1 && err != nil {
		t.Errorf("BulkConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = BulkConnectNodes([]string{"2"}, []string{"3"}, file)
	if count != 1 && err != nil {
		t.Errorf("BulkConnectNodes() inserted %d,%q but expected 1,nil", count, err.Error())
	}
}
