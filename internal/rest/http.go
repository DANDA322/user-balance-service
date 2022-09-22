package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DANDA322/user-balance-service/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

type Balance interface {
	GetBalance(ctx context.Context, accountID int, currency string) (float64, error)
	AddDepositToWallet(ctx context.Context, accountID int, transaction models.Transaction) error
	WithdrawMoneyFromWallet(ctx context.Context, accountID int, transaction models.Transaction) error
	TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error
}

func NewRouter(log *logrus.Logger, balance Balance) chi.Router {
	handler := newHandler(log, balance)
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.AllowAll().Handler)
	r.NotFound(notFoundHandler)
	r.Use(handler.auth)
	r.Get("/test", handler.Test)
	r.Get("/getBalance", handler.GetBalance)
	r.Post("/addDeposit", handler.DepositMoneyToWallet)
	r.Post("/withdrawMoney", handler.WithdrawMoneyFromWallet)

	return r
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *handler) Test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{\"response\":\"test!\"}"))
}

func (h *handler) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Errorf("unable to encode data %v", err)
	}
}

func (h *handler) writeErrResponse(w http.ResponseWriter, code int, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if newErr := json.NewEncoder(w).Encode(map[string]interface{}{"error": err}); newErr != nil {
		h.log.Errorf("unable to encode %v", newErr)
	}
}
