package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vctrl/currency-service/currency/internal/clients/currency"
	"github.com/vctrl/currency-service/currency/internal/dto"
	"github.com/vctrl/currency-service/currency/internal/repository"

	"go.uber.org/zap"
)

type Currency struct {
	currencyRepo repository.Currency
	client       currency.Currency
	logger       *zap.Logger
}

func NewCurrency(
	repo repository.Currency,
	client currency.Currency,
	logger *zap.Logger,
) Currency {
	return Currency{
		currencyRepo: repo,
		client:       client,
		logger:       logger,
	}
}

func (s *Currency) GetCurrencyRatesInInterval(ctx context.Context, reqDTO *dto.CurrencyRequestDTO) ([]repository.CurrencyRate, error) {
	reqDTO.TargetCurrency = strings.ToLower(reqDTO.TargetCurrency)
	rates, err := s.currencyRepo.FindInInterval(ctx, reqDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch currency rates in interval: %w", err)
	}

	return rates, nil
}

func (s *Currency) FetchAndSaveCurrencyRates(ctx context.Context, baseCurrency string) error {
	rates, err := s.client.FetchCurrentRates(ctx, baseCurrency)
	if err != nil {
		return fmt.Errorf("client.FetchCurrentRates: %s", err)
	}

	date, err := time.Parse("2006-01-02", rates.Date)
	if err != nil {
		return fmt.Errorf("Failed to parse currency date: %v ", err)
	}

	if err := s.currencyRepo.Save(ctx, date, baseCurrency, rates.Rub); err != nil { // todo want to pass struct
		return fmt.Errorf("Failed to save currency rates: %v ", err)
	}

	s.logger.Info("Currency rates fetched and saved", zap.Any("rates", rates))
	return nil
}

func (s *Currency) UpdateCurrencyRate(ctx context.Context, reqDTO *dto.UpdateCurrencyRequestDTO) error {
	reqDTO.TargetCurrency = strings.ToLower(reqDTO.TargetCurrency)
	_, err := s.currencyRepo.UpdateRate(ctx, reqDTO)
	if err != nil {
		return fmt.Errorf("failed to update currency rate: %w", err)
	}

	s.logger.Info(
		"Currency rate updated",
		zap.Any("currency", reqDTO.BaseCurrency+reqDTO.TargetCurrency),
		zap.Any("rate", reqDTO.RateRecord),
	)

	return nil
}
