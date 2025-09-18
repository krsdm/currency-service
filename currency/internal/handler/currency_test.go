package handler

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vctrl/currency-service/currency/internal/dto"
	"github.com/vctrl/currency-service/currency/internal/handler/mocks"
	"github.com/vctrl/currency-service/currency/internal/repository"
	"github.com/vctrl/currency-service/pkg/currency"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_request_count",
			Help: "Test count",
		},
		[]string{"method"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_request_duration",
			Help:    "Test duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
	appUptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_app_uptime",
		Help: "Test app uptime",
	})
)

func TestGetRate(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := mocks.NewMockCurrencyService(ctrl)

	service.EXPECT().
		GetCurrencyRatesInInterval(context.Background(), &dto.CurrencyRequestDTO{BaseCurrency: "RUB"}).
		Return([]repository.CurrencyRate{}, nil)

	logger := zaptest.NewLogger(t)
	server := NewCurrencyServer(service,
		logger,
		requestCount,
		requestDuration,
		appUptime,
	)

	expected := &currency.GetRateResponse{
		Rates: make([]*currency.RateRecord, 0),
	}

	ctx := context.Background()
	req := &currency.GetRateRequest{
		DateFrom: timestamppb.New(time.Time{}.UTC()),
		DateTo:   timestamppb.New(time.Time{}.UTC()),
	}

	fact, err := server.GetRate(ctx, req)

	require.NoError(t, err)

	assert.Equal(t, expected, fact)
}

func TestUpdateRate(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := mocks.NewMockCurrencyService(ctrl)
	testDTO := &dto.UpdateCurrencyRequestDTO{
		BaseCurrency:   "RUB",
		TargetCurrency: "USD",
		RateRecord: dto.RateRecordDTO{
			Date: time.Time{},
			Rate: 10,
		},
	}

	service.EXPECT().
		UpdateCurrencyRate(context.Background(), testDTO).
		Return(nil)

	server := NewCurrencyServer(service,
		zap.NewNop(),
		requestCount,
		requestDuration,
		appUptime,
	)

	_, err := server.UpdateRate(context.Background(), &currency.UpdateRateRequest{
		Currency: "USD",
		RateRecord: &currency.RateRecord{
			Date: timestamppb.New(time.Time{}.UTC()),
			Rate: 10,
		},
	})

	require.NoError(t, err)
}
