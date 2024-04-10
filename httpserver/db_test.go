package main_test

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/tursodatabase/go-libsql"
	"testing"
)

func TestDB(t *testing.T) {
	ctx := context.TODO()

	db, err := sql.Open("libsql", "file:./db.sqlite")
	if err != nil {
		return
	}
	defer db.Close()

	db.Exec("CREATE TABLE hello IF NOT EXISTS()")

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Print("success")
}
