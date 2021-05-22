package simplegraph

import "testing"

func TestResolveDbFileReference(t *testing.T) {
	path := "/tmp/database.sqlite?_foreign_keys=true"
	actualPath, actualErr := resolveDbFileReference("/tmp", "database.sqlite")
	if actualPath != path {
		t.Errorf("resolveDbFileReference(\"/tmp\", \"database.sqlite\" = %q but expected %q", actualPath, path)
	}
	if actualErr != nil {
		t.Errorf("resolveDbFileReference(\"/tmp\", \"database.sqlite\" = %q but expected nil", actualErr)
	}
}
