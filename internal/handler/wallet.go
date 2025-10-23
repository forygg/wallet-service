package handler

import (
	"encoding/json"
	"net/http"
	"fmt"

	"wallet-service/internal/model"
	"wallet-service/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	walletService service.WalletServiceInterface // Используем интерфейс
}

func NewWalletHandler(walletService service.WalletServiceInterface) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

func (h *WalletHandler) ProcessOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.WalletOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateOperationRequest(req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	operation := model.WalletOperation(req)

	if err := h.walletService.ProcessOperation(r.Context(), operation); err != nil {
		switch err {
		case service.ErrWalletNotFound:
			respondWithError(w, http.StatusNotFound, "Wallet not found")
		case service.ErrInsufficientFunds:
			respondWithError(w, http.StatusBadRequest, "Insufficient funds")
		case service.ErrOptimisticLock:
			respondWithError(w, http.StatusConflict, "Operation conflict, please retry")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	vars := mux.Vars(r)
	walletIDStr := vars["walletId"]

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	balance, err := h.walletService.GetBalance(r.Context(), walletID)
	if err != nil {
		if err == service.ErrWalletNotFound {
			respondWithError(w, http.StatusNotFound, "Wallet not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response := model.BalanceResponse{
		WalletID: walletID,
		Balance:  balance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateOperationRequest(req model.WalletOperationRequest) error {
	if req.WalletID == uuid.Nil {
		return fmt.Errorf("walletId is required")
	}

	if req.OperationType != model.OperationTypeDeposit && req.OperationType != model.OperationTypeWithdraw {
		return fmt.Errorf("operationType must be DEPOSIT or WITHDRAW")
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}

	return nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}