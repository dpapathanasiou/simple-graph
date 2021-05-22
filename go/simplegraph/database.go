package simplegraph

import (
	"database/sql"
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
)

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
