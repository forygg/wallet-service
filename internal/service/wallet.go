package service

import (
	"context"
	"errors"
	"time"

	"wallet-service/internal/model"
	"wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Экспортируем ошибки для использования в хендлерах
var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrOptimisticLock    = errors.New("optimistic lock conflict")
)

type WalletService struct {
	repo    repository.WalletRepository
	retries int
}

func NewWalletService(repo repository.WalletRepository, retries int) *WalletService {
	return &WalletService{
		repo:    repo,
		retries: retries,
	}
}

func (s *WalletService) GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
	balance, err := s.repo.GetBalance(ctx, id)
	if errors.Is(err, repository.ErrWalletNotFound) {
		return decimal.Zero, ErrWalletNotFound
	}
	return balance, err
}

func (s *WalletService) ProcessOperation(ctx context.Context, op model.WalletOperation) error {
	for i := 0; i < s.retries; i++ {
		err := s.repo.UpdateBalance(ctx, op)
		if errors.Is(err, repository.ErrOptimisticLock) {
			time.Sleep(time.Duration(i*i) * time.Millisecond * 10)
			continue
		}
		
		// Маппим ошибки репозитория на ошибки сервиса
		if errors.Is(err, repository.ErrWalletNotFound) {
			return ErrWalletNotFound
		}
		if errors.Is(err, repository.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}
		return err
	}
	return ErrOptimisticLock
}