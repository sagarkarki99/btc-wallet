package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Connect() {
	ctx := context.Background()
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	DB = db
	if err != nil {
		panic(err)
	}
	if err = db.DB.Ping(); err != nil {
		fmt.Println("error connecting to db:", err)
		return
	}
	fmt.Println("Database connected!!")
}
