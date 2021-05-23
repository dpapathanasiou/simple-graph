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

func TestGenerateSearchStatement(t *testing.T) {
	where := generateSearchEquals(map[string]string{"name": "Steve"})
	single := "json_extract(body, '$.name') = ?"
	if where != single {
		t.Errorf("generateSearchEquals() = %q but expected %q", where, single)
	}

	props := map[string]string{"name": "Steve", "type": "founder"}
	where = generateSearchLike(props)
	multiple := "json_extract(body, '$.name') LIKE ? AND json_extract(body, '$.type') LIKE ?"
	if where != multiple {
		t.Errorf("generateSearchLike() = %q but expected %q", where, multiple)
	}

	where = generateSearchStatement(props, true)
	sql := "SELECT body FROM nodes WHERE json_extract(body, '$.name') = ? AND json_extract(body, '$.type') = ?"
	if where != sql {
		t.Errorf("generateSearchStatement() = %q but expected %q", where, sql)
	}

	where = generateSearchStatement(props, false)
	sql = "SELECT body FROM nodes WHERE json_extract(body, '$.name') LIKE ? AND json_extract(body, '$.type') LIKE ?"
	if where != sql {
		t.Errorf("generateSearchStatement() = %q but expected %q", where, sql)
	}
}

func TestInitializeAndCrud(t *testing.T) {
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

	apple := `{"id":"1","name":"Apple Computer Company","type":["company","start-up"],"founded":"April 1, 1976"}`
	count, err := AddNode([]byte(apple), file)
	if count != 1 && err != nil {
		t.Errorf("AddNode() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	woz := `{"name": "Steve Wozniak", "type":["person","engineer","founder"]}`
	count, err = AddNodeAndId([]byte(woz), "2", file)
	if count != 1 && err != nil {
		t.Errorf("AddNodeAndId() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodes("1", "2", file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodes() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode([]byte(apple), file)
	if count != 0 && !ErrorMatches(err, UNIQUE_ID_CONSTRAINT) {
		t.Errorf("AddNode() inserted %d,%q but expected 0,%q", count, err.Error(), UNIQUE_ID_CONSTRAINT)
	}

	count, err = AddNode([]byte(woz), file)
	if count != 0 && !ErrorMatches(err, ID_CONSTRAINT) {
		t.Errorf("AddNode() inserted %d,%q but expected 0,%q", count, err.Error(), ID_CONSTRAINT)
	}

	node, err := FindNode("1", file)
	if node != apple && err != nil {
		t.Errorf("FindNode() produced %q,%q but expected %q,nil", node, err.Error(), apple)
	}

	node, err = FindNode("3", file)
	if node != "" && !ErrorMatches(err, NO_ROWS_FOUND) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", NO_ROWS_FOUND)
	}

	if !RemoveNode("2", file) {
		t.Error("RemoveNode() returned false but expected true")
	}

	node, err = FindNode("2", file)
	if node != "" && !ErrorMatches(err, NO_ROWS_FOUND) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", NO_ROWS_FOUND)
	}
}
