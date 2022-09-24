package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DANDA322/user-balance-service/internal/models"
	"github.com/stretchr/testify/require"
)

const (
	token1 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ" //nolint:lll,gosec
	token2 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjozMzMsInJvbGUiOiJhZG1pbiJ9.SxZaQcVFDVSb72MPMLbGPE5-23s-FZO9Lgip2oS13vwKy9f5Qe0L_xrCtWQbrAodlFphwmF-dCTd59hAaahcoNzN1Jgj0b15NJBKDcQgZDhQN8jehXrDrFdfj2UUi9y3KpHfRtepBDPiMXNCUd5zaY_3eW5ilbBtUj8GDchN0SiRyg9d3v4THvk21o3CDWRwLe8exKTdP7KCfuGeqLG8315aMSIuOUCNw25m-JKzQUYlgeaxQDK0d6DDitogBy0WYI77KZXVK5M5r-tYWj9aIcy7pCk2jCZ-NkuL5ekLbYfI5NHzNbF3sJUdxE4GkIx2x4LrX38lJvZe80bH0aQIMQ" //nolint:lll,gosec
)

var transaction1 = &models.Transaction{
	Amount:  100.50,
	Comment: "Пополнение баланса",
}

var transaction2 = &models.Transaction{
	Amount:  100.50,
	Comment: "Снятие средств",
}

var transferTransaction = &models.TransferTransaction{
	Target:  333,
	Amount:  100.5,
	Comment: "Перевод",
}

var transaction3 = &models.Transaction{
	Amount:  10000,
	Comment: "Снятие средств",
}

var transaction4 = &models.Transaction{
	Amount:  50,
	Comment: "Пополнение баланса",
}

var transaction5 = &models.Transaction{
	Amount:  1000.50,
	Comment: "Пополнение баланса",
}

var balance1 = &models.Balance{
	Currency: "RUB",
	Amount:   100.5,
}

var balance2 = &models.Balance{
	Currency: "USD",
	Amount:   1.7286,
}

var transferBalance0 = &models.Balance{
	Currency: "RUB",
	Amount:   0,
}

var transferBalance1 = &models.Balance{
	Currency: "RUB",
	Amount:   201,
}

var transactions = []models.TransactionFullInfo{
	{
		ID:             1,
		WalletID:       nil,
		Amount:         100.5,
		TargetWalletID: 1,
		Comment:        "Пополнение баланса",
		Timestamp:      time.Now(),
	},
	{
		ID:             2,
		WalletID:       nil,
		Amount:         1000.5,
		TargetWalletID: 1,
		Comment:        "Пополнение баланса",
		Timestamp:      time.Now(),
	},
	{
		ID:             3,
		WalletID:       nil,
		Amount:         -100.5,
		TargetWalletID: 1,
		Comment:        "Снятие средств",
		Timestamp:      time.Now(),
	},
}

func (s *IntegrationTestSuite) TestAddDeposit() {
	resp, code, err := s.processRequest("POST", "/wallet/addDeposit", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, balance1)
}

func (s *IntegrationTestSuite) TestGetBalance() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance1.Amount, respStruct.Amount)
	require.Equal(s.T(), balance1.Currency, respStruct.Currency)
}

func (s *IntegrationTestSuite) TestGetBalanceUSD() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("GET", "/wallet/getBalance?currency=USD", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance2.Amount, respStruct.Amount)
	require.Equal(s.T(), balance2.Currency, respStruct.Currency)
}

func (s *IntegrationTestSuite) TestGetBalanceInvalidCurrency() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("GET", "/wallet/getBalance?currency=awda", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, code)
	require.Equal(s.T(), "{\"error\":\"invalid currency symbols\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalanceNotFound() {
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalanceNotAuth() {
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", "", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusUnauthorized, code)
	require.Equal(s.T(), "{\"error\":\"Unauthorized\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token1, transaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, transferBalance0)
}

func (s *IntegrationTestSuite) TestWithdrawMoneyNotFound() {
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token1, transaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoneyNotEnoughMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token1, transaction3)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough money on the balance\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestTransferMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	depositMoney(s.T(), s, token2, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/transferMoney", token1, transferTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, transferBalance0)
	checkBalance(s.T(), s, token2, transferBalance1)
}

func (s *IntegrationTestSuite) TestTransferMoneyWalletNotFound() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/transferMoney", token1, transferTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestTransferMoneyNotEnoughMoney() {
	depositMoney(s.T(), s, token1, transaction4)
	depositMoney(s.T(), s, token2, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/transferMoney", token1, transferTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough money on the balance\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetWalletTransactions() {
	depositMoney(s.T(), s, token1, transaction1)
	depositMoney(s.T(), s, token1, transaction5)
	withdrawMoney(s.T(), s, token1, transaction2)
	resp, code, err := s.processRequest("GET", "/wallet/getTransactions", token1, nil)
	var respStruct []models.TransactionFullInfo
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), transactions, respStruct)
}

func depositMoney(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.Transaction) {
	t.Helper()
	resp, code, err := s.processRequest("POST", "/wallet/addDeposit", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func withdrawMoney(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.Transaction) {
	t.Helper()
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func checkBalance(t *testing.T, s *IntegrationTestSuite, token string, balance *models.Balance) {
	t.Helper()
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", token, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance.Amount, respStruct.Amount)
	require.Equal(s.T(), balance.Currency, respStruct.Currency)
}
