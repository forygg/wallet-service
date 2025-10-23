package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func main() {
	baseURL := "http://localhost:8080"
	walletID := uuid.New()
	concurrentRequests := 1000
	var successCount int32
	var errorCount int32

	fmt.Printf("Creating wallet: %s\n", walletID)
	
	// Сначала создаем кошелек через DEPOSIT
	if err := createWallet(baseURL, walletID); err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	fmt.Printf("Starting load test: %d concurrent requests to wallet %s\n", concurrentRequests, walletID)

	var wg sync.WaitGroup
	start := time.Now()

	// Запускаем 1000 concurrent запросов
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			if err := sendDepositRequest(baseURL, walletID, decimal.NewFromInt(1)); err != nil {
				atomic.AddInt32(&errorCount, 1)
				fmt.Printf("Request %d failed: %v\n", index, err)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n=== LOAD TEST RESULTS ===\n")
	fmt.Printf("Total requests: %d\n", concurrentRequests)
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("RPS: %.2f\n", float64(concurrentRequests)/duration.Seconds())
	
	// Проверяем финальный баланс
	balance, err := getBalance(baseURL, walletID)
	if err != nil {
		log.Printf("Failed to get final balance: %v", err)
	} else {
		expected := concurrentRequests + 1
		fmt.Printf("Final balance: %s (expected: %d)\n", balance.String(), expected)
		fmt.Printf("Balance correct: %t\n", balance.Equal(decimal.NewFromInt(int64(expected))))
	}
}

func createWallet(baseURL string, walletID uuid.UUID) error {
	return sendDepositRequest(baseURL, walletID, decimal.NewFromInt(1))
}

func sendDepositRequest(baseURL string, walletID uuid.UUID, amount decimal.Decimal) error {
	reqBody := struct {
		WalletID      string `json:"walletId"`
		OperationType string `json:"operationType"`
		Amount        string `json:"amount"`
	}{
		WalletID:      walletID.String(),
		OperationType: "DEPOSIT",
		Amount:        amount.String(),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	fmt.Printf("Sending request: %s\n", string(body))

	resp, err := http.Post(baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return fmt.Errorf("unexpected status: %d, body: %v", resp.StatusCode, errorBody)
	}

	return nil
}

func getBalance(baseURL string, walletID uuid.UUID) (decimal.Decimal, error) {
	resp, err := http.Get(baseURL + "/api/v1/wallets/" + walletID.String())
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var balanceResp struct {
		Balance decimal.Decimal `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return decimal.Zero, err
	}

	return balanceResp.Balance, nil
}