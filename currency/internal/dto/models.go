package dto

import (
	"time"

	"github.com/vctrl/currency-service/pkg/currency"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CurrencyRequestDTO struct {
	BaseCurrency   string
	TargetCurrency string
	DateFrom       time.Time
	DateTo         time.Time
}

type UpdateCurrencyRequestDTO struct {
	BaseCurrency   string
	TargetCurrency string
	RateRecord     RateRecordDTO
}

type CurrencyResponseDTO struct {
	Currency string
	Rates    []RateRecordDTO
}

type RateRecordDTO struct {
	Date time.Time
	Rate float32
}

const (
	DefaultBaseCurrency = "RUB"
)

func CurrencyRequestDTOFromProtobuf(req *currency.GetRateRequest, baseCurrency string) *CurrencyRequestDTO {
	return &CurrencyRequestDTO{
		BaseCurrency:   baseCurrency,
		TargetCurrency: req.Currency,
		DateFrom:       req.DateFrom.AsTime(),
		DateTo:         req.DateTo.AsTime(),
	}
}

func (dto *CurrencyResponseDTO) ToProtobuf() *currency.GetRateResponse {
	rateRecords := make([]*currency.RateRecord, 0, len(dto.Rates))
	for _, record := range dto.Rates {
		rateRecords = append(
			rateRecords, &currency.RateRecord{
				Date: timestamppb.New(record.Date),
				Rate: record.Rate,
			},
		)
	}

	return &currency.GetRateResponse{
		Currency: dto.Currency,
		Rates:    rateRecords,
	}
}

func UpdateCurrencyRequestDTOFromProtobuf(req *currency.UpdateRateRequest, baseCurrency string) *UpdateCurrencyRequestDTO {
	return &UpdateCurrencyRequestDTO{
		BaseCurrency:   baseCurrency,
		TargetCurrency: req.Currency,
		RateRecord: RateRecordDTO{
			Date: req.RateRecord.Date.AsTime(),
			Rate: req.RateRecord.Rate,
		},
	}
}
