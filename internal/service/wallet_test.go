package service

import (
	"context"
	"testing"

	"wallet-service/internal/model"
	"wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock репозитория
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockWalletRepository) UpdateBalance(ctx context.Context, op model.WalletOperation) error {
	args := m.Called(ctx, op)
	return args.Error(0)
}

func TestWalletService_GetBalance(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	service := NewWalletService(mockRepo, 3)

	walletID := uuid.New()
	expectedBalance := decimal.NewFromInt(1000)

	// Настраиваем mock
	mockRepo.On("GetBalance", mock.Anything, walletID).Return(expectedBalance, nil)

	// Вызываем метод
	balance, err := service.GetBalance(context.Background(), walletID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.True(t, expectedBalance.Equal(balance))
	mockRepo.AssertExpectations(t)
}

func TestWalletService_GetBalance_WalletNotFound(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	service := NewWalletService(mockRepo, 3)

	walletID := uuid.New()

	// Настраиваем mock для возврата ошибки
	mockRepo.On("GetBalance", mock.Anything, walletID).Return(decimal.Zero, repository.ErrWalletNotFound)

	// Вызываем метод
	balance, err := service.GetBalance(context.Background(), walletID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Equal(t, ErrWalletNotFound, err)
	assert.True(t, decimal.Zero.Equal(balance))
	mockRepo.AssertExpectations(t)
}

func TestWalletService_ProcessOperation_Deposit(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	service := NewWalletService(mockRepo, 3)

	walletID := uuid.New()
	operation := model.WalletOperation{
		WalletID:      walletID,
		OperationType: model.OperationTypeDeposit,
		Amount:        decimal.NewFromInt(500),
	}

	// Настраиваем mock
	mockRepo.On("UpdateBalance", mock.Anything, operation).Return(nil)

	// Вызываем метод
	err := service.ProcessOperation(context.Background(), operation)

	// Проверяем результат
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWalletService_ProcessOperation_OptimisticLockRetry(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	service := NewWalletService(mockRepo, 3)

	walletID := uuid.New()
	operation := model.WalletOperation{
		WalletID:      walletID,
		OperationType: model.OperationTypeDeposit,
		Amount:        decimal.NewFromInt(500),
	}

	// Настраиваем mock: первые 2 вызова - ошибка, третий - успех
	mockRepo.On("UpdateBalance", mock.Anything, operation).Return(repository.ErrOptimisticLock).Twice()
	mockRepo.On("UpdateBalance", mock.Anything, operation).Return(nil).Once()

	// Вызываем метод
	err := service.ProcessOperation(context.Background(), operation)

	// Проверяем результат
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "UpdateBalance", 3)
}