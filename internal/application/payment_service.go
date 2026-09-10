package application

import (
	"context"
	"nabutilivanie/internal/domain"
)

type PaymentService struct{ payments PaymentRepository }

func NewPaymentService(payments PaymentRepository) *PaymentService {
	return &PaymentService{payments: payments}
}
func (s *PaymentService) Process(ctx context.Context, provider, externalID string, telegramID, bottles int64, payload any) (bool, domain.Player, error) {
	return s.payments.ProcessPayment(ctx, provider, externalID, telegramID, bottles, payload)
}
