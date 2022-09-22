package rest

import (
	"crypto/rsa"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	sessionInfo := ctx.Value("sessionInfo").(models.SessionInfo)
	accountID := sessionInfo.AccountID
	balance, err := h.balance.GetBalance(ctx, accountID, currency)
	switch {
	case err == nil:
	case errors.Is(err, sql.ErrNoRows):
		h.writeErrResponse(w, http.StatusNotFound, "Not found")
		return
	default:
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
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value("sessionInfo").(models.SessionInfo)
	accountID := sessionInfo.AccountID
	err := h.balance.AddDepositToWallet(ctx, accountID, transaction)
	if err != nil {
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) WithdrawMoneyFromWallet(w http.ResponseWriter, r *http.Request) {
	transaction := models.Transaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value("sessionInfo").(models.SessionInfo)
	accountID := sessionInfo.AccountID
	err := h.balance.WithdrawMoneyFromWallet(ctx, accountID, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrNotEnoughMoney):
		h.writeErrResponse(w, http.StatusConflict, err)
		return
	default:
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

// func (h *handler) TransferMoney(w http.ResponseWriter, r *http.Request) {
//	transaction := models.TransferTransaction{}
//	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
//		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
//		h.log.Info(err)
//		return
//	}
//	ctx := r.Context()
//	sessionInfo := ctx.Value("sessionInfo").(models.SessionInfo)
//	accountID := sessionInfo.AccountId
//
//	err := h.balance.TransferMoney(ctx, accountID, transaction)
// }
