package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	currencyClient "github.com/vctrl/currency-service/currency/internal/clients/currency"
	"github.com/vctrl/currency-service/currency/internal/dto"
	"github.com/vctrl/currency-service/currency/internal/repository/mocks"
	"go.uber.org/zap"
)

type CurrencySuite struct {
	suite.Suite
	ctrl                   *gomock.Controller
	currencyRepositoryMock *mocks.MockCurrencyRepository
	currencyService        Currency
}

func TestCurrencySuite(t *testing.T) {
	suite.Run(t, new(CurrencySuite))
}

func (s *CurrencySuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.currencyRepositoryMock = mocks.NewMockCurrencyRepository(s.ctrl)
	client := currencyClient.Currency{}
	s.currencyService = NewCurrency(s.currencyRepositoryMock, client, zap.NewNop())
}

func (s *CurrencySuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CurrencySuite) TestUpdateCurrencyRate_CorrectDTO() {
	updateDTO := &dto.UpdateCurrencyRequestDTO{
		BaseCurrency:   "RUB",
		TargetCurrency: "usd",
		RateRecord: dto.RateRecordDTO{
			Date: time.Time{},
			Rate: 10,
		},
	}

	s.currencyRepositoryMock.EXPECT().
		UpdateRate(gomock.Any(), updateDTO).
		Return(int64(0), nil)

	err := s.currencyService.UpdateCurrencyRate(context.Background(), updateDTO)
	s.Require().NoError(err)
}

func (s *CurrencySuite) TestUpdateCurrencyRate_MissingTargetCurrency() {
	updateDTO := &dto.UpdateCurrencyRequestDTO{
		BaseCurrency:   "RUB",
		TargetCurrency: "",
		RateRecord:     dto.RateRecordDTO{},
	}

	err := s.currencyService.UpdateCurrencyRate(context.Background(), updateDTO)
	s.Require().Error(err)
}

func (s *CurrencySuite) TestUpdateCurrencyRate_WrongRateRecord() {
	updateDTO := &dto.UpdateCurrencyRequestDTO{
		BaseCurrency:   "RUB",
		TargetCurrency: "USD",
		RateRecord: dto.RateRecordDTO{
			Date: time.Time{},
			Rate: -10,
		},
	}

	err := s.currencyService.UpdateCurrencyRate(context.Background(), updateDTO)
	s.Require().Error(err)
}
