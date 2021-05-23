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

	count, err = AddNodeAndId([]byte(`{"name": "Steve Wozniak", "type":["person","engineer","founder"]}`), "2", file)
	if count != 1 && err != nil {
		t.Errorf("AddNodeAndId() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = ConnectNodes("1", "2", file)
	if count != 1 && err != nil {
		t.Errorf("ConnectNodes() inserted %d,%q but expected 1,nil", count, err.Error())
	}

	count, err = AddNode([]byte(apple), file)
	uniqueIdConstraint := "UNIQUE constraint failed: nodes.id"
	if count != 0 && !ErrorMatches(err, uniqueIdConstraint) {
		t.Errorf("AddNode() inserted %d,%q but expected 0,%q", count, err.Error(), uniqueIdConstraint)
	}

	node, err := FindNode("1", file)
	if node != apple && err != nil {
		t.Errorf("FindNode() produced %q,%q but expected %q,nil", node, err.Error(), apple)
	}

	node, err = FindNode("3", file)
	notFound := "sql: no rows in result set"
	if node != "" && !ErrorMatches(err, notFound) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", notFound)
	}

	if !RemoveNode("2", file) {
		t.Error("RemoveNode() returned false but expected true")
	}

	node, err = FindNode("2", file)
	if node != "" && !ErrorMatches(err, notFound) {
		t.Errorf("FindNode() produced %q,%q but expected %q,%q", node, err.Error(), "", notFound)
	}
}
