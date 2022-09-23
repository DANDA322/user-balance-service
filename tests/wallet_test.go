package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DANDA322/user-balance-service/internal/models"
	"github.com/stretchr/testify/require"
)

const (
	token1 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50SWQiOjU1NSwicm9sZSI6ImFkbWluIn0.NSoip83iAGX2NHMNBTIpyaqcNlvuuqO-QRkr9JibJOQkuc0X1ep5dfDS_GsFOUxllVAZgMaIin3cLc3dXfkkfhzVj6MpsSh4a6HXODVPpfhvkP8aJi_wUo03D2Jp9yTOe2QxXiAfnIrXDkfRx90bhGEPX79qbbLQDbntCiWesgWchuL916TpdYVaHoOyS_oHHcM6TKPsOhkDxe-3M8BVmwSQbizFbjw_KE6wfbFUznA4xQe4Z62idID9qZSXxt_ILN_lgzzUJfHFCmmWN1LKAdCPxNPhfzc9HsIUVhlO3Mxm5lN5UDCPpFz-ArDQ4y-bTKx05v9YrWkb1aqiz8h38w" //nolint:lll,gosec
)

var transaction1 = &models.Transaction{
	Amount:  100.50,
	Comment: "Пополнение баланса",
}

var transaction2 = &models.Transaction{
	Amount:  100.50,
	Comment: "Снятие средств",
}

var transaction3 = &models.Transaction{
	Amount:  10000,
	Comment: "Снятие средств",
}

var balance1 = &models.Balance{
	Currency: "RUB",
	Amount:   100.5,
}

var balance2 = &models.Balance{
	Currency: "USD",
	Amount:   1.7365395,
}

func (s *IntegrationTestSuite) TestAddDeposit() {
	resp, code, err := s.processRequest("POST", "/wallet/addDeposit", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalance() {
	addDeposit(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance1.Amount, respStruct.Amount)
	require.Equal(s.T(), balance1.Currency, respStruct.Currency)
}

func (s *IntegrationTestSuite) TestGetBalanceUSD() {
	addDeposit(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("GET", "/wallet/getBalance?currency=USD", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance2.Amount, respStruct.Amount)
	require.Equal(s.T(), balance2.Currency, respStruct.Currency)
}

func (s *IntegrationTestSuite) TestGetBalanceInvalidCurrency() {
	addDeposit(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("GET", "/wallet/getBalance?currency=awda", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, code)
	require.Equal(s.T(), "{\"error\":\"invalid currency symbols\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalanceNotFound() {
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"Not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalanceNotAuth() {
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", "", transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusUnauthorized, code)
	require.Equal(s.T(), "{\"error\":\"Unauthorized\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoney() {
	addDeposit(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token1, transaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoneyNotFound() {
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token1, transaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"Not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoneyNotEnoughMoney() {
	addDeposit(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest("POST", "/wallet/withdrawMoney", token1, transaction3)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough money on the balance\"}\n", string(resp))
}

func addDeposit(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.Transaction) {
	t.Helper()
	resp, code, err := s.processRequest("POST", "/wallet/addDeposit", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}
