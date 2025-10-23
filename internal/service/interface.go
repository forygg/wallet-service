package service

import (
	"context"

	"wallet-service/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WalletServiceInterface определяет контракт сервиса для использования в хендлерах
type WalletServiceInterface interface {
	GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, error)
	ProcessOperation(ctx context.Context, op model.WalletOperation) error
}