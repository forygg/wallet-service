package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wallet-service/internal/model"
	"wallet-service/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/gorilla/mux"
)

// Mock сервиса для интеграционных тестов
type MockWalletService struct{}

func (m *MockWalletService) GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
	if id == uuid.MustParse("123e4567-e89b-12d3-a456-426614174000") {
		return decimal.NewFromInt(1000), nil
	}
	return decimal.Zero, service.ErrWalletNotFound
}

func (m *MockWalletService) ProcessOperation(ctx context.Context, op model.WalletOperation) error {
	if op.WalletID == uuid.MustParse("00000000-0000-0000-0000-000000000000") {
		return service.ErrWalletNotFound
	}
	if op.OperationType == model.OperationTypeWithdraw && op.Amount.GreaterThan(decimal.NewFromInt(1000)) {
		return service.ErrInsufficientFunds
	}
	return nil
}

func TestWalletHandler_ProcessOperation_Success(t *testing.T) {
	service := &MockWalletService{}
	handler := NewWalletHandler(service)

	// Подготавливаем запрос
	reqBody := model.WalletOperationRequest{
		WalletID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		OperationType: model.OperationTypeDeposit,
		Amount:        decimal.NewFromInt(500),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Вызываем хендлер
	rr := httptest.NewRecorder()
	handler.ProcessOperation(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])
}

func TestWalletHandler_GetBalance_Success(t *testing.T) {
	service := &MockWalletService{}
	handler := NewWalletHandler(service)

	// Создаем роутер и регистрируем хендлер
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetBalance).Methods("GET")

	// Подготавливаем запрос с правильным URL
	req := httptest.NewRequest("GET", "/api/v1/wallets/123e4567-e89b-12d3-a456-426614174000", nil)

	// Вызываем через роутер
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response model.BalanceResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", response.WalletID.String())
	assert.True(t, decimal.NewFromInt(1000).Equal(response.Balance))
}

func TestWalletHandler_GetBalance_NotFound(t *testing.T) {
	service := &MockWalletService{}
	handler := NewWalletHandler(service)

	// Создаем роутер и регистрируем хендлер
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetBalance).Methods("GET")

	// Подготавливаем запрос с несуществующим ID
	req := httptest.NewRequest("GET", "/api/v1/wallets/00000000-0000-0000-0000-000000000000", nil)

	// Вызываем через роутер
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestWalletHandler_ProcessOperation_InvalidMethod(t *testing.T) {
	service := &MockWalletService{}
	handler := NewWalletHandler(service)

	// Создаем роутер
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet", handler.ProcessOperation).Methods("POST")

	// Пробуем GET вместо POST
	req := httptest.NewRequest("GET", "/api/v1/wallet", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestWalletHandler_ProcessOperation_InvalidUUID(t *testing.T) {
	service := &MockWalletService{}
	handler := NewWalletHandler(service)

	// Создаем роутер
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetBalance).Methods("GET")

	// Запрос с невалидным UUID
	req := httptest.NewRequest("GET", "/api/v1/wallets/invalid-uuid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}