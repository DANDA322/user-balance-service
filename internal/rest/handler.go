package rest

import "github.com/sirupsen/logrus"

type handler struct {
	log     *logrus.Logger
	balance Balance
}

func newHandler(log *logrus.Logger, balance Balance) *handler {
	return &handler{
		log:     log,
		balance: balance,
	}
}
