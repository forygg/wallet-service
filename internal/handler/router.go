package handler

import (
	"net/http"

	"wallet-service/internal/service"
	"github.com/gorilla/mux"
)

func NewRouter(walletService *service.WalletService) http.Handler {
	router := mux.NewRouter()
	walletHandler := NewWalletHandler(walletService)

	router.HandleFunc("/api/v1/wallet", walletHandler.ProcessOperation).Methods("POST")
	router.HandleFunc("/api/v1/wallets/{walletId}", walletHandler.GetBalance).Methods("GET")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return router
}