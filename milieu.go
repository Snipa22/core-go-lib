package corelib

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Milieu struct {
	Pgx         *pgxpool.Pool
	Redis       *redis.Client
	transaction pgx.Tx
	psqlconn    *pgxpool.Conn
}

var bg = context.Background()

func (c *Milieu) GetTransaction() (pgx.Tx, error) {
	var err error
	if c.transaction == nil {
		if c.psqlconn == nil {
			if c.psqlconn, err = c.Pgx.Acquire(bg); err != nil {
				return c.transaction, err
			}
		}
		if c.transaction, err = c.psqlconn.Begin(bg); err != nil {
			return c.transaction, err
		}
	}
	return c.transaction, nil
}

func (c *Milieu) Cleanup() {
	if c.transaction != nil {
		c.transaction.Rollback(context.Background())
	}
	if c.psqlconn != nil {
		c.psqlconn.Release()
	}
}
