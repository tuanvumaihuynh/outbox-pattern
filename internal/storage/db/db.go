package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row

	CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults

	// WithTx executes a function in a new transaction.
	WithTx(ctx context.Context, txFunc func(DB) error) error
}

type HealthChecker interface {
	IsHealthy(ctx context.Context) (bool, error)
}

var (
	_ DB            = (*Client)(nil)
	_ HealthChecker = (*Client)(nil)
)

type Client struct {
	*pgxpool.Pool
}

// NewClient creates a new db client.
func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool}
}

func (p *Client) WithTx(ctx context.Context, txFunc func(DB) error) (err error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rbErr := tx.Rollback(ctx)
			if !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = errors.Join(err, rbErr)
			}
		}
	}()

	txDB := &txWrapper{Tx: tx}
	if err = txFunc(txDB); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		err = fmt.Errorf("commit transaction: %w", err)
	}

	return err
}

func (p *Client) IsHealthy(ctx context.Context) (bool, error) {
	err := p.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("ping database: %w", err)
	}
	return true, nil
}

type txWrapper struct {
	pgx.Tx
}

func (t *txWrapper) WithTx(_ context.Context, txFunc func(DB) error) error {
	return txFunc(t)
}
