package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/vctrl/currency-service/currency/internal/dto"
	"github.com/vctrl/currency-service/pkg/currency"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s CurrencyServer) GetRate(ctx context.Context, request *currency.GetRateRequest) (*currency.GetRateResponse, error) {
	start := time.Now()
	reqDTO := dto.CurrencyRequestDTOFromProtobuf(request, dto.DefaultBaseCurrency)

	// todo метрики в мидлвары
	s.requestCount.WithLabelValues("GetRate").Inc()
	rates, err := s.service.GetCurrencyRatesInInterval(ctx, reqDTO)
	if err != nil {
		return nil, fmt.Errorf("service.GetCurrencyRatesInInterval: %w", err)
	}

	rateRecords := make([]*currency.RateRecord, len(rates))
	for i, rate := range rates {
		rateRecords[i] = &currency.RateRecord{
			Date: timestamppb.New(rate.Date),
			Rate: rate.Rate,
		}
	}

	s.requestDuration.WithLabelValues("GetExchangeRate").Observe(time.Since(start).Seconds())
	return &currency.GetRateResponse{
		Currency: reqDTO.TargetCurrency,
		Rates:    rateRecords,
	}, nil
}

func (s CurrencyServer) UpdateRate(ctx context.Context, request *currency.UpdateRateRequest) (*emptypb.Empty, error) {
	reqDto := dto.UpdateCurrencyRequestDTOFromProtobuf(request, dto.DefaultBaseCurrency)
	err := s.service.UpdateCurrencyRate(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateRate: %w", err)
	}

	return &emptypb.Empty{}, nil
}
