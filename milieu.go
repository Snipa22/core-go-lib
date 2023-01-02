// Package corelib provides a number of support functions that I utilize in a number of my regular applications
// It's generally not designed for public consumption, but it's out there in case it's desired
// This is opinionated towards Gin as I use it for most of my web-based systems.
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

// GetTransaction will return the current transaction if one is generated in this particular Milieu instance
// otherwise it will try to use a conn attached to the instance, failing that, it will go all the way and generate both
// a dedicated sql conn, as required for the txn, and the txn itself
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

// Cleanup force rollbacks the transaction stored in the current Milieu instance
func (c *Milieu) Cleanup() {
	if c.transaction != nil {
		_ = c.transaction.Rollback(bg)
	}
	if c.psqlconn != nil {
		c.psqlconn.Release()
	}
}

// Clone allocates a new Milieu instance, cloning the data from the parent, but clearing any transaction and conn
func (c *Milieu) Clone() Milieu {
	return Milieu{
		Pgx:   c.Pgx,
		Redis: c.Redis,
	}
}

// NewMilieu issues a standard Milieu object
// psqlURI is a standard PSQL URI string in the format postgres://user:pass@host:port/db
// redisURI is a standard URI string in the format: redis://user:pass@host:port/db
func NewMilieu(psqlURI string, redisURI string) (Milieu, error) {
	pgpool, err := pgxpool.Connect(bg, psqlURI)
	if err != nil {
		return Milieu{}, err
	}
	opts, err := redis.ParseURL(redisURI)
	if err != nil {
		return Milieu{}, err
	}
	rdb := redis.NewClient(opts)
	return Milieu{
		Pgx:   pgpool,
		Redis: rdb,
	}, nil
}
