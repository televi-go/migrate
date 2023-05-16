package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slices"
	"log"
	"testing"
)

func TestSort(t *testing.T) {
	var strs = []string{
		"AB.x",
		"AB-2.x",
	}
	slices.SortFunc(strs, func(a, b string) bool {
		return a < b
	})
	fmt.Printf("%#v\n", strs)
}

//go:embed all:migrations/*
var xfs embed.FS

func TestRunInEmbed(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		log.Fatalln(err)
	}

	err = RunInEmbed(db, "migrations", xfs, MysqlDialect)
	if err != nil {
		log.Fatalln(err)
	}

	row := db.QueryRow("select x from a limit 1")
	var x string
	err = row.Scan(&x)
	if x != "x" {
		log.Fatalln("Not working")
	}
}
