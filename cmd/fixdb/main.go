package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5434/gocode?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE schema_migrations SET version = 3, dirty = false")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migration version forced to 3 and dirty flag reset!")
}
