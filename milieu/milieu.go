// Package corelib provides a number of support functions that I utilize in a number of my regular applications
// It's generally not designed for public consumption, but it's out there in case it's desired
// This is opinionated towards Gin as I use it for most of my web-based systems.
package milieu

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type Milieu struct {
	pgx         *pgxpool.Pool
	redis       *redis.Client
	transaction pgx.Tx
	psqlconn    *pgxpool.Conn
	logger      *logrus.Logger
	logEntry    *logrus.Entry
	sentry      bool
}

var bg = context.Background()

// GetTransaction will return the current transaction if one is generated in this particular Milieu instance
// otherwise it will try to use a conn attached to the instance, failing that, it will go all the way and generate both
// a dedicated sql conn, as required for the txn, and the txn itself
func (c *Milieu) GetTransaction() (pgx.Tx, error) {
	if c.pgx == nil {
		return nil, ErrPSQLNotActive
	}
	var err error
	if c.transaction == nil {
		if c.psqlconn == nil {
			if c.psqlconn, err = c.pgx.Acquire(bg); err != nil {
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
// This also wipes the logger clean, so any fields will need to be created/added once more.
func (c *Milieu) Clone() Milieu {
	return Milieu{
		pgx:      c.pgx,
		redis:    c.redis,
		logger:   c.logger,
		logEntry: logrus.NewEntry(c.logger),
		sentry:   c.sentry,
	}
}

// AddLoggerField adds a log field to the internal log entry for this instance
func (c *Milieu) AddLoggerField(fieldName string, fieldValue interface{}) {
	c.logEntry = c.logEntry.WithField(fieldName, fieldValue)
}

// SetLogLevel overrides the global log level for the internal logger in Milieu
func (c *Milieu) SetLogLevel(level logrus.Level) {
	c.logger.SetLevel(level)
}

func (c *Milieu) Debug(msg string) {
	c.logEntry.Debug(msg)
}
func (c *Milieu) Trace(msg string) {
	c.logEntry.Trace(msg)
}
func (c *Milieu) Info(msg string) {
	c.logEntry.Info(msg)
}
func (c *Milieu) Warn(msg string) {
	c.logEntry.Warn(msg)
}
func (c *Milieu) Error(msg string) {
	c.logEntry.Error(msg)
}
func (c *Milieu) Fatal(msg string) {
	c.logEntry.Fatal(msg)
}
func (c *Milieu) Panic(msg string) {
	c.logEntry.Panic(msg)
}

// CaptureException wraps the default sentry.CaptureException call, allowing us to check to see if Milieu has been
// configured to use sentry yet, if not, we handle this gracefully
func (c *Milieu) CaptureException(err error) {
	if c.sentry == true {
		sentry.CaptureException(err)
	} else {
		c.Error(err.Error())
	}
}

func (c *Milieu) GetRawPGXPool() *pgxpool.Pool {
	return c.pgx
}

// NewMilieu issues a standard Milieu object
// psqlURI is a standard PSQL URI string in the format postgres://user:pass@host:port/db
// redisURI is a standard URI string in the format: redis://user:pass@host:port/db
func NewMilieu(psqlURI *string, redisURI *string, sentryDSN *string) (*Milieu, error) {
	internalLogger := logrus.New()
	intMilieu := &Milieu{
		pgx:      nil,
		redis:    nil,
		logger:   internalLogger,
		logEntry: logrus.NewEntry(internalLogger),
		sentry:   false,
	}
	if sentryDSN != nil {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: *sentryDSN,
		}); err != nil {
			return intMilieu, err
		} else {
			intMilieu.sentry = true
		}
	}
	if psqlURI != nil {
		pgpool, err := pgxpool.Connect(bg, *psqlURI)
		if err != nil {
			return intMilieu, err
		}
		intMilieu.pgx = pgpool
	}
	if redisURI != nil {
		opts, err := redis.ParseURL(*redisURI)
		if err != nil {
			return intMilieu, err
		}
		intMilieu.redis = redis.NewClient(opts)
	}
	return intMilieu, nil
}
