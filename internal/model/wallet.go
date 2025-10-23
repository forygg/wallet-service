package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Wallet struct {
    ID      uuid.UUID       `json:"id" db:"id"`
    Balance decimal.Decimal `json:"balance" db:"balance"`
    Version int             `json:"version" db:"version"`
}

type OperationType string

const (
    OperationTypeDeposit  OperationType = "DEPOSIT"
    OperationTypeWithdraw OperationType = "WITHDRAW"
)

type WalletOperation struct {
    WalletID      uuid.UUID       `json:"walletId"`
    OperationType OperationType   `json:"operationType"`
    Amount        decimal.Decimal `json:"amount"`
}