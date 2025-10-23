package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletOperationRequest struct {
    WalletID      uuid.UUID       `json:"walletId"`
    OperationType OperationType   `json:"operationType"`
    Amount        decimal.Decimal `json:"amount"`
}

type BalanceResponse struct {
    WalletID uuid.UUID       `json:"walletId"`
    Balance  decimal.Decimal `json:"balance"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}