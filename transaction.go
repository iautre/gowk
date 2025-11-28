package gowk

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"sync"
)

const (
	TRANSACTION = "transaction"
)

type Transaction struct {
	Begin bool
	mu    sync.Mutex
	Tx    pgx.Tx
}

func TransactionHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(TRANSACTION, &Transaction{})
		defer func() {
			if tx, ok := ctx.Value(TRANSACTION).(*Transaction); ok && tx != nil && tx.Begin && tx.Tx != nil {
				if err := recover(); err != nil {
					Rollback(ctx)
					panic(err)
				} else if len(ctx.Errors) > 0 {
					Rollback(ctx)
				} else {
					Commit(ctx)
				}
			}
		}()
		ctx.Next()
	}
}

func Begin(ctx context.Context) error {
	tx := ctx.Value(TRANSACTION).(*Transaction)
	if tx == nil {
		slog.InfoContext(ctx, "开启事务失败，未配置事务")
		return errors.New("开启事务失败，未配置事务 TransactionHandler")
	}
	tx.mu.Lock()
	defer tx.mu.Unlock()
	tx.Begin = true
	return nil
}

func End(ctx context.Context, err error) {
	if err != nil {
		Rollback(ctx)
	} else {
		Commit(ctx)
	}
}

func Commit(ctx context.Context) {
	tx := ctx.Value(TRANSACTION).(*Transaction)
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.Begin && tx.Tx != nil {
		if err := tx.Tx.Commit(ctx); err != nil {
			slog.ErrorContext(ctx, "事务提交失败", err.Error())
		}
		tx.Begin = false
		tx.Tx = nil
	}
}

func Rollback(ctx context.Context) {
	tx := ctx.Value(TRANSACTION).(*Transaction)
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.Begin && tx.Tx != nil {
		if err := tx.Tx.Rollback(ctx); err != nil {
			slog.ErrorContext(ctx, "事务回滚失败", err.Error())
		}
		tx.Begin = false
		tx.Tx = nil
	}
}
