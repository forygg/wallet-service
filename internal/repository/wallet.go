package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"wallet-service/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrWalletNotFound		= errors.New("wallet not found")
	ErrOptimisticLock 		= errors.New("optimistic lock conflict")
	ErrInsufficientFunds	= errors.New("insufficient funds")
)

type WalletRepository interface {
    GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, error)
    UpdateBalance(ctx context.Context, op model.WalletOperation) error
}

type walletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
	var balance decimal.Decimal
	
	query := `SELECT balance FROM wallets WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return decimal.Zero, ErrWalletNotFound
		}
		return decimal.Zero, fmt.Errorf("failed to get balance: %w", err)
	}
	
	return balance, nil
}


func (r *walletRepository) UpdateBalance(ctx context.Context, op model.WalletOperation) error {
    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Пытаемся найти кошелек
    var currentBalance decimal.Decimal
    var version int
    
    query := `SELECT balance, version FROM wallets WHERE id = $1 FOR UPDATE`
    err = tx.QueryRowContext(ctx, query, op.WalletID).Scan(&currentBalance, &version)
    
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return fmt.Errorf("failed to get wallet: %w", err)
    }

    // Если кошелек не найден
    if errors.Is(err, sql.ErrNoRows) {
        // Для DEPOSIT - создаем новый кошелек
        if op.OperationType == model.OperationTypeDeposit {
            createQuery := `INSERT INTO wallets (id, balance, version) VALUES ($1, $2, $3)`
            _, err := tx.ExecContext(ctx, createQuery, op.WalletID, op.Amount, 1)
            if err != nil {
                return fmt.Errorf("failed to create wallet: %w", err)
            }
            return tx.Commit()
        } else {
            // Для WITHDRAW - кошелек не существует
            return ErrWalletNotFound
        }
    }

    // Если кошелек существует - обычная логика
    if op.OperationType == model.OperationTypeWithdraw {
        if currentBalance.LessThan(op.Amount) {
            return ErrInsufficientFunds
        }
    }

    var newBalance decimal.Decimal
    if op.OperationType == model.OperationTypeDeposit {
        newBalance = currentBalance.Add(op.Amount)
    } else {
        newBalance = currentBalance.Sub(op.Amount)
    }

    updateQuery := `UPDATE wallets SET balance = $1, version = version + 1 WHERE id = $2 AND version = $3`
    result, err := tx.ExecContext(ctx, updateQuery, newBalance, op.WalletID, version)
    if err != nil {
        return fmt.Errorf("failed to update balance: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return ErrOptimisticLock
    }

    return tx.Commit()
}

