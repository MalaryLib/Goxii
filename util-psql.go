package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func GenerateConnStr(user, password, host, db string) (string) {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s", user, password, host, db,
	)
}

func GetDatabaseConn(connStr string) (*sql.DB, error) {
	return sql.Open("postgres", connStr)
}