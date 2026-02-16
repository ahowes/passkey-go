package db

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/ahowes/passkey-go/models"
)

func Connect(dsn, user, password string) *bun.DB {
	opts := []pgdriver.Option{pgdriver.WithDSN(dsn)}
	if user != "" {
		opts = append(opts, pgdriver.WithUser(user))
	}
	if password != "" {
		opts = append(opts, pgdriver.WithPassword(password))
	}
	sqldb := sql.OpenDB(pgdriver.NewConnector(opts...))
	return bun.NewDB(sqldb, pgdialect.New())
}

func CreateTables(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*models.User)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		Model((*models.WebAuthnCredential)(nil)).
		IfNotExists().
		ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`).
		Exec(ctx)
	return err
}
