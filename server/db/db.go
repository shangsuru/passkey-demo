package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"
)

func GetDB() *bun.DB {
	dbString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		panic(err)
	}

	return bun.NewDB(db, pgdialect.New())
}

func GetTestDB() *bun.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	testDB := bun.NewDB(db, sqlitedialect.New())

	// Run migrations
	ctx := context.Background()
	migrations := migrate.NewMigrations()
	if err = migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
	migrator := migrate.NewMigrator(testDB, migrations)
	if err = migrator.Init(ctx); err != nil {
		panic(err)
	}
	_, err = migrator.Migrate(ctx)
	if err != nil {
		panic(err)
	}

	return testDB
}
