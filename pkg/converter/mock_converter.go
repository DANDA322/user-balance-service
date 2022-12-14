package converter

import (
	"context"

	"github.com/DANDA322/user-balance-service/internal/models"
)

type MockConverter struct {
}

func NewMockConverter() *MockConverter {
	return &MockConverter{}
}

func (c *MockConverter) GetRate(ctx context.Context, currency string) (float64, error) {
	if currency == "USD" {
		return 0.0172, nil
	}
	return 0, models.ErrInvalidCurrencySymbols
}
