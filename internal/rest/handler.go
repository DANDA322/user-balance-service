package rest

import (
	"crypto/rsa"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DANDA322/user-balance-service/internal/models"
	"github.com/sirupsen/logrus"
)

//go:embed public.pub
var publicSigningKey []byte

type handler struct {
	log     *logrus.Logger
	balance Balance
	pubKey  *rsa.PublicKey
}

func newHandler(log *logrus.Logger, balance Balance) *handler {
	return &handler{
		log:     log,
		balance: balance,
		pubKey:  mustGetPublicKey(publicSigningKey),
	}
}

func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	currency := r.URL.Query().Get("currency")
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	accountID := sessionInfo.AccountID
	balance, err := h.balance.GetBalance(ctx, accountID, currency)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrInvalidCurrencySymbols):
		h.writeErrResponse(w, http.StatusBadRequest, models.ErrInvalidCurrencySymbols.Error())
		return
	default:
		h.log.Errorf("Error get balance: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	if currency == "" {
		currency = "RUB"
	}
	result := models.Balance{
		Currency: currency,
		Amount:   balance,
	}
	h.writeJSONResponse(w, result)
}

func (h *handler) DepositMoneyToWallet(w http.ResponseWriter, r *http.Request) {
	transaction := models.Transaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	accountID := sessionInfo.AccountID
	err := h.balance.AddDepositToWallet(ctx, accountID, transaction)
	switch {
	case err == nil:
	default:
		h.log.Errorf("Error deposit money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) WithdrawMoneyFromWallet(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	transaction := models.Transaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	accountID := sessionInfo.AccountID
	err := h.balance.WithdrawMoneyFromWallet(ctx, accountID, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughMoney.Error())
		return
	default:
		h.log.Errorf("Error withdraw money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) TransferMoney(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	transaction := models.TransferTransaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	accountID := sessionInfo.AccountID
	err := h.balance.TransferMoney(ctx, accountID, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughMoney.Error())
		return
	default:
		h.log.Errorf("Error transfer money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	accountID := sessionInfo.AccountID
	from, err := h.parseTime(r.URL.Query().Get("from"))
	if err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't parse time")
		h.log.Info(err)
		return
	}
	to, err := h.parseTime(r.URL.Query().Get("to"))
	if err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't parse time")
		h.log.Info(to)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		h.writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		h.writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	sorting := r.URL.Query().Get("sorting")
	descending := r.URL.Query().Get("descending")
	var transactions []models.TransactionFullInfo
	queryParams := models.TransactionsQueryParams{
		From:       from,
		To:         to,
		Limit:      limit,
		Offset:     offset,
		Sorting:    sorting,
		Descending: descending,
	}
	transactions, err = h.balance.GetWalletTransaction(ctx, accountID, &queryParams)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	default:
		h.log.Errorf("Error get wallet transaction: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, transactions)
}

const dateTimeFmt = "2006-01-02T15:04:05Z"

func (h *handler) parseTime(s string) (time.Time, error) {
	t, err := time.Parse(dateTimeFmt, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("could nor parse time: %w", err)
	}
	return t, err
}
