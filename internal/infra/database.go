package infra

import (
	"context"
	"database/sql"
	"fmt"

	"xadrez/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func OpenDatabase(ctx context.Context, dsn string) (*ent.Client, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir o banco de dados: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("falha ao criar o esquema do banco de dados: %w", err)
	}
	return client, nil
}
