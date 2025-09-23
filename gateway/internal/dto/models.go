package dto

import "time"

type ParsedCurrencyRequest struct {
	Currency string
	DateFrom time.Time
	DateTo   time.Time
}

type CurrencyResponse struct {
	Currency string
	Rates    []CurrencyRate
}

type CurrencyRate struct {
	Rate float32
	Date time.Time
}

type RegisterRequest struct {
	Username string
	Password string
}

type RegisterResponse struct {
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
}

type ParsedUpdateRateRequest struct {
	Currency string
	Rate     float32
	Date     time.Time
}
