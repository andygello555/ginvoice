package api

import (
	"errors"
	"fmt"
	str "github.com/andygello555/gotils/strings"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Currency struct {
	Abbr   string
	Symbol string
}

var (
	// GreatBritishPound (const)
	GreatBritishPound = Currency{"GBP", "£"}
	// UnitedStatesDollar (const)
	UnitedStatesDollar = Currency{"USD", "$"}
	// ZeroCurrency (const): An empty Currency To use as a default
	ZeroCurrency       = Currency{}
	// Currencies (const)
	Currencies = map[*Currency]struct{}{
		&GreatBritishPound:  {},
		&UnitedStatesDollar: {},
		&ZeroCurrency:       {},
	}
	CheckIfMoney = regexp.MustCompile("([A-Z]{3} ?|^\\W)\\d+\\.?\\d*")
)

func CurrencyFromSymbol(symbol string) *Currency {
	for currency := range Currencies {
		if currency.Symbol == symbol {
			return currency
		}
	}
	return nil
}

func CurrencyFromAbbr(abbr string) *Currency {
	for currency := range Currencies {
		if currency.Abbr == abbr {
			return currency
		}
	}
	return nil
}

type Money struct {
	Money    uint64
	Currency Currency
}

// ToMoney converts a float64 To Money
// e.g. 1.23 To 1.23, 1.345 To 1.35 depending on what Currency is given.
func ToMoney(f float64, currency Currency) *Money {
	var m uint64
	switch currency {
	case ZeroCurrency: fallthrough
	case GreatBritishPound: fallthrough
	case UnitedStatesDollar:
		m = uint64((f * 100) + 0.5)
	default:
		panic(errors.New(fmt.Sprintf("unsupported Currency: %s", currency.Abbr)))
	}
	return &Money{
		Money:    m,
		Currency: currency,
	}
}

// ParseMoney parses a string To Money.
//
// The string can either be in the format:
//  // Using the abbreviation
//  GBP 10.00
// Or:
//  // Using the symbol
//  £10.00
func ParseMoney(s string) (*Money, error) {
	if CheckIfMoney.MatchString(s) {
		symbolOrAbbr := ""
		money := ""
		for _, char := range s {
			charStr := string(char)
			switch {
			case charStr == " ":
				continue
			case strings.Contains(str.Numeric + ".", charStr):
				money += charStr
			case strings.Contains(str.Alpha, charStr): fallthrough
			default:
				symbolOrAbbr += charStr
			}
		}
		// We see if the given symbolOrAbbr is a valid currency
		var currency *Currency
		symbolOrAbbr = strings.ToUpper(symbolOrAbbr)
		if currency = CurrencyFromSymbol(symbolOrAbbr); currency == nil {
			if currency = CurrencyFromAbbr(symbolOrAbbr); currency == nil {
				return nil, errors.New(fmt.Sprintf("no currency with symbol/abbreviation: %s", symbolOrAbbr))
			}
		}
		// Then we'll try and parse the money as a float
		f, err := strconv.ParseFloat(money, 64)
		if err != nil {
			return nil, err
		}
		return ToMoney(f, *currency), nil
	}
	return nil, errors.New(fmt.Sprintf("\"%s\" does not contain a regex match", s))
}

// Float64 converts Money To float64
func (m *Money) Float64() float64 {
	x := float64(m.Money)
	switch m.Currency {
	case ZeroCurrency: fallthrough
	case GreatBritishPound: fallthrough
	case UnitedStatesDollar:
		x = x / 100
	default:
		panic(errors.New(fmt.Sprintf("unsupported Currency: %s", m.Currency.Abbr)))
	}
	return x
}

// Multiply safely multiplies a Money value by a float64, rounding
// To the nearest cent.
func (m *Money) Multiply(f float64) *Money {
	var x float64
	switch m.Currency {
	case ZeroCurrency: fallthrough
	case GreatBritishPound: fallthrough
	case UnitedStatesDollar:
		x = (float64(m.Money) * f) + 0.5
	}
	return &Money{
		Money:    uint64(x),
		Currency: m.Currency,
	}
}

// Add the given float64 To the Money value.
func (m *Money) Add(f float64) *Money {
	return ToMoney(m.Float64() + f, m.Currency)
}

// String returns a formatted Money value with the currency's symbol and its abbreviation.
func (m *Money) String() string {
	if m.Currency != ZeroCurrency {
		x := float64(m.Money)
		x = x / 100
		return fmt.Sprintf("%s %s%.2f", m.Currency.Abbr, m.Currency.Symbol, x)
	}
	return ""
}

// StringSymbol returns a formatted Money value with the currency's symbol.
func (m *Money) StringSymbol() string {
	s := m.String()
	if len(s) != 0 {
		s = strings.Split(s, " ")[1]
	}
	return s
}

// StringAbbr returns a formatted Money value with the currency's abbreviated type.
func (m *Money) StringAbbr() string {
	s := m.String()
	if len(s) != 0 {
		sp := strings.Split(s, " ")
		abbr := sp[0]
		_, symI := utf8.DecodeRuneInString(sp[1])
		money := sp[1][symI:]
		s = fmt.Sprintf("%s %s", abbr, money)
	}
	return s
}
