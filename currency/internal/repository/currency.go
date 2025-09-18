package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/vctrl/currency-service/currency/internal/dto"

	_ "github.com/lib/pq"
)

type Currency struct {
	DB *sql.DB
}

type CurrencyRate struct {
	Date time.Time
	Rate float32
}

func NewCurrency(db *sql.DB) (Currency, error) {
	return Currency{
		DB: db,
	}, nil
}

func (repo *Currency) Save(
	ctx context.Context,
	date time.Time,
	baseCurrency string,
	rates map[string]float64,
) error {
	ratesJSON, err := json.Marshal(rates)
	if err != nil {
		return fmt.Errorf("failed to marshal currency rates: %w", err)
	}

	_, err = repo.DB.ExecContext(
		ctx,
		`INSERT INTO exchange_rates (date, base_currency, currency_rates) VALUES ($1, $2, $3)`,
		date, baseCurrency, ratesJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to save exchange rates: %w", err)
	}
	return nil
}

func (repo *Currency) FindInInterval(
	ctx context.Context,
	dto *dto.CurrencyRequestDTO,
) ([]CurrencyRate, error) {
	query := `
		SELECT date, (currency_rates ->> $1)::float 
		FROM exchange_rates
		WHERE date::date BETWEEN $2 AND $3 AND base_currency = $4
	`

	rows, err := repo.DB.QueryContext(
		ctx,
		query,
		dto.TargetCurrency,
		dto.DateFrom.Format("2006-01-02"),
		dto.DateTo.Format("2006-01-02"),
		dto.BaseCurrency,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query exchange rates: %w", err)
	}
	defer rows.Close()

	var rates []CurrencyRate
	for rows.Next() {
		var rate CurrencyRate
		if err := rows.Scan(&rate.Date, &rate.Rate); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return rates, nil
}

func (repo *Currency) UpdateRate(ctx context.Context, dto *dto.UpdateCurrencyRequestDTO) (int64, error) {
	query := `
		UPDATE exchange_rates
		SET currency_rates = jsonb_set(currency_rates, $1, to_jsonb($2::numeric), true)
		WHERE date::date = $3 AND base_currency = $4;
	`

	result, err := repo.DB.ExecContext(
		ctx,
		query,
		pq.Array([]string{dto.TargetCurrency}),
		dto.RateRecord.Rate,
		dto.RateRecord.Date,
		dto.BaseCurrency,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to update exchange rate: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}
