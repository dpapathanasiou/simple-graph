package simplegraph

import (
	"os"
	"testing"
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

func TestGenerateSearchStatement(t *testing.T) {
	where := generateSearchEquals(map[string]string{"name": "Steve"})
	single := "json_extract(body, '$.name') = ?"
	if where != single {
		t.Errorf("generateSearchEquals() = %q but expected %q", where, single)
	}

	props := map[string]string{"name": "Steve", "type": "founder"}
	where = generateSearchLike(props)
	multiple := "json_extract(body, '$.name') LIKE ? AND json_extract(body, '$.type') LIKE ?"
	multipleReversed := "json_extract(body, '$.type') LIKE ? AND json_extract(body, '$.name') LIKE ?"
	if where != multiple && where != multipleReversed {
		t.Errorf("generateSearchLike() = %q but expected %q", where, multiple)
	}

	where = generateSearchStatement(props, true)
	sql := "SELECT body FROM nodes WHERE json_extract(body, '$.name') = ? AND json_extract(body, '$.type') = ?"
	sqlReversed := "SELECT body FROM nodes WHERE json_extract(body, '$.type') = ? AND json_extract(body, '$.name') = ?"
	if where != sql && where != sqlReversed {
		t.Errorf("generateSearchStatement() = %q but expected %q", where, sql)
	}

	where = generateSearchStatement(props, false)
	sql = "SELECT body FROM nodes WHERE json_extract(body, '$.name') LIKE ? AND json_extract(body, '$.type') LIKE ?"
	sqlReversed = "SELECT body FROM nodes WHERE json_extract(body, '$.type') LIKE ? AND json_extract(body, '$.name') LIKE ?"
	if where != sql && where != sqlReversed {
		t.Errorf("generateSearchStatement() = %q but expected %q", where, sql)
	}

	equality := []string{"Steve", "founder"}
	for _, binding := range generateSearchBindings(props, false, false) {
		if !arrayContains(equality, binding) {
			t.Errorf("generateSearchBindings() was missing %q", binding)
		}
	}

	startsWith := []string{"Steve%", "founder%"}
	for _, binding := range generateSearchBindings(props, true, false) {
		if !arrayContains(startsWith, binding) {
			t.Errorf("generateSearchBindings() was missing %q", binding)
		}
	}

	contains := []string{"%Steve%", "%founder%"}
	for _, binding := range generateSearchBindings(props, false, true) {
		if !arrayContains(contains, binding) {
			t.Errorf("generateSearchBindings() was missing %q", binding)
		}
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
	expected := []string{"3", "1", `{"action":"founded"}`, "4", "1", `{}`}
	for i, actual := range makeBulkEdgeInserts([]string{"3", "4"}, []string{"1", "1"}, []string{`{"action":"founded"}`, `{}`}) {
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

	apple := `{"name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976"}`
	count, err := AddNode("1", []byte(apple), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	woz := `{"id":"2","name":"Steve Wozniak","type":["person","engineer","founder"]}`
	count, err = AddNode("2", []byte(woz), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	jobs := `{"id":"3","name":"Steve Jobs","type":["person","designer","founder"]}`
	count, err = AddNode("3", []byte(jobs), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("4", []byte(`{"name": "Ronald Wayne", "type":["person","administrator","founder"]}`), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode("5", []byte(`{"name": "Mike Markkula", "type":["person","investor"]}`), file)
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

	founded := `{"action": "founded"}`
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

	count, err = ConnectNodesWithProperties("5", "1", []byte(`{"action": "invested", "equity": 80000, "debt": 170000}`), file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodesWithProperties("1", "4", []byte(`{"action": "divested", "amount": 800, "date": "April 12, 1976"}`), file)
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

	nodes, err := FindNodes(map[string]string{"name": "Steve"}, true, false, file)
	if err != nil {
		t.Errorf("FindNodes() produced an error %s but expected nil", err.Error())
	}
	if !arrayContains(nodes, woz) {
		t.Errorf("FindNodes() did not return %s as expected", woz)
	}
	if !arrayContains(nodes, jobs) {
		t.Errorf("FindNodes() did not return %s as expected", jobs)
	}

	wozNick := `{"name":"Steve Wozniak","type":["person","engineer","founder"],"nickname":"Woz","id":"2"}`
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

	idList, traverseErr := TraverseFromTo("2", "3", Traverse, file)
	if traverseErr != nil {
		t.Errorf("TraverseFromTo() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"2", "1", "3"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFromTo() did not return %s as expected", expectedId)
		}
	}

	idList, traverseErr = TraverseFrom("5", TraverseInbound, file)
	if traverseErr != nil {
		t.Errorf("TraverseFrom() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"5"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFrom() did not return %s as expected", expectedId)
		}
	}

	idList, traverseErr = TraverseFrom("5", TraverseOutbound, file)
	if traverseErr != nil {
		t.Errorf("TraverseFrom() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"5", "1", "4"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFrom() did not return %s as expected", expectedId)
		}
	}

	idList, traverseErr = TraverseFrom("5", Traverse, file)
	if traverseErr != nil {
		t.Errorf("TraverseFrom() produced an error %s but expected nil", traverseErr.Error())
	}
	for _, expectedId := range []string{"5", "1", "2", "3", "4"} {
		if !arrayContains(idList, expectedId) {
			t.Errorf("TraverseFrom() did not return %s as expected", expectedId)
		}
	}

	edges, err := ConnectionsIn("1", file)
	if err != nil {
		t.Errorf("ConnectionsIn() produced an error %s but expected nil", err.Error())
	}
	expected := []EdgeData{{"1", "4", `{"action":"divested","amount":800,"date":"April 12, 1976"}`}}
	for i, exp := range expected {
		if edges[i] != exp {
			t.Errorf("ConnectionsIn() produced %q but expected %q", edges[i], exp)
		}
	}

	edges, err = ConnectionsOut("1", file)
	if err != nil {
		t.Errorf("ConnectionsIn() produced an error %s but expected nil", err.Error())
	}
	expected = []EdgeData{{"2", "1", `{"action":"founded"}`},
		{"3", "1", `{"action":"founded"}`},
		{"4", "1", `{"action":"founded"}`},
		{"5", "1", `{"action":"invested","equity":80000,"debt":170000}`}}
	for i, exp := range expected {
		if edges[i] != exp {
			t.Errorf("ConnectionsOut() produced %q but expected %q", edges[i], exp)
		}
	}

	edges, err = Connections("1", file)
	if err != nil {
		t.Errorf("Connections() produced an error %s but expected nil", err.Error())
	}
	expected = []EdgeData{{"1", "4", `{"action":"divested","amount":800,"date":"April 12, 1976"}`},
		{"2", "1", `{"action":"founded"}`},
		{"3", "1", `{"action":"founded"}`},
		{"4", "1", `{"action":"founded"}`},
		{"5", "1", `{"action":"invested","equity":80000,"debt":170000}`}}
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

	nodes := [][]byte{[]byte(`{"name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976"}`),
		[]byte(`{"id":"2","name":"Steve Wozniak","type":["person","engineer","founder"]}`),
		[]byte(`{"id":"3","name":"Steve Jobs","type":["person","designer","founder"]}`),
		[]byte(`{"name": "Ronald Wayne", "type":["person","administrator","founder"]}`),
		[]byte(`{"name": "Mike Markkula", "type":["person","investor"]}`)}
	ids := []string{"1", "2", "3", "4", "5"}

	count, err := AddNodes(ids, nodes, file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	sources := []string{"2", "3", "4", "5", "1"}
	targets := []string{"1", "1", "1", "1", "4"}
	properties := []string{`{"action": "founded"}`,
		`{"action": "founded"}`,
		`{"action": "founded"}`,
		`{"action": "invested", "equity": 80000, "debt": 170000}`,
		`{"action": "divested", "amount": 800, "date": "April 12, 1976"}`}

	count, err = BulkConnectNodesWithProperties(sources, targets, properties, file)
	if count != 1 && err != nil {
		t.Errorf("BulkConnectNodesWithProperties() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = BulkConnectNodes([]string{"2"}, []string{"3"}, file)
	if count != 1 && err != nil {
		t.Errorf("BulkConnectNodes() inserted %d,%q but expected 1,nil", count, err.Error())
	}
}
