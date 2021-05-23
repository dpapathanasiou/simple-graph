package simplegraph

import (
	"os"
	"strings"
	"testing"
)

func ErrorMatches(actual error, expected string) bool {
	if actual == nil {
		return expected == ""
	}
	if expected == "" {
		return false
	}
	return strings.Contains(actual.Error(), expected)
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
	count := AddNode([]byte(apple), file)
	if count != 1 {
		t.Errorf("AddNode() inserted %d but expected 1", count)
	}

	count = AddNodeAndId([]byte(`{"name": "Steve Wozniak", "type":["person","engineer","founder"]}`), "2", file)
	if count != 1 {
		t.Errorf("AddNodeAndId() inserted %d but expected 1", count)
	}

	count = ConnectNodes("1", "2", file)
	if count != 1 {
		t.Errorf("ConnectNodes() inserted %d but expected 1", count)
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
