package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Connect() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

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
