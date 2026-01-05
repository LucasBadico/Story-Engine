package world

import (
	"errors"
)

// TimeConfig represents the calendar/time configuration for a world
type TimeConfig struct {
	// Unidade base do temporal_offset (para display)
	BaseUnit string `json:"base_unit"` // "year", "day", "hour", "custom"

	// Estrutura do calendário
	HoursPerDay   float64 `json:"hours_per_day"`   // Default: 24
	DaysPerWeek   int     `json:"days_per_week"`   // Default: 7
	DaysPerYear   int     `json:"days_per_year"`   // Default: 365
	MonthsPerYear int     `json:"months_per_year"` // Default: 12

	// Meses com duração variável (opcional)
	MonthLengths []int `json:"month_lengths,omitempty"` // [31, 28, 31, 30, ...]

	// Nomes customizados (opcional)
	MonthNames []string `json:"month_names,omitempty"` // ["Janeiro", "Fevereiro", ...]
	DayNames   []string `json:"day_names,omitempty"`   // ["Domingo", "Segunda", ...]

	// Época/Era inicial (opcional)
	EraName  string `json:"era_name,omitempty"`  // "Era Comum", "Terceira Era"
	YearZero int    `json:"year_zero,omitempty"` // Ano inicial da timeline
}

// DefaultTimeConfig returns the default time configuration (Earth-like)
func DefaultTimeConfig() *TimeConfig {
	return &TimeConfig{
		BaseUnit:      "year",
		HoursPerDay:   24,
		DaysPerWeek:   7,
		DaysPerYear:   365,
		MonthsPerYear: 12,
		MonthLengths:  []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31},
	}
}

// DaysInMonth returns the number of days in a given month (1-indexed)
func (tc *TimeConfig) DaysInMonth(month int) int {
	if month < 1 || month > tc.MonthsPerYear {
		return 0
	}
	if len(tc.MonthLengths) > 0 && month <= len(tc.MonthLengths) {
		return tc.MonthLengths[month-1]
	}
	// Default: equal months
	return tc.DaysPerYear / tc.MonthsPerYear
}

// Validate validates the time configuration
func (tc *TimeConfig) Validate() error {
	if tc.HoursPerDay <= 0 {
		return ErrInvalidTimeConfig
	}
	if tc.DaysPerWeek <= 0 {
		return ErrInvalidTimeConfig
	}
	if tc.DaysPerYear <= 0 {
		return ErrInvalidTimeConfig
	}
	if tc.MonthsPerYear <= 0 {
		return ErrInvalidTimeConfig
	}
	if len(tc.MonthLengths) > 0 {
		sum := 0
		for _, days := range tc.MonthLengths {
			sum += days
		}
		if sum != tc.DaysPerYear {
			return ErrInvalidTimeConfig
		}
	}
	return nil
}

var (
	ErrInvalidTimeConfig = errors.New("invalid time configuration")
)

