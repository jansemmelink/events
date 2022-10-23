package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-msvc/errors"
)

type Amount struct {
	Currency *Currency
	Cents    int
}

func (a Amount) String() string {
	s := ""
	if a.Cents < 100 || a.Cents%100 != 0 {
		s = fmt.Sprintf("%d.%02d", a.Cents/100, a.Cents%100)
	} else {
		s = fmt.Sprintf("%d", a.Cents/100)
	}
	if a.Currency == nil {
		return s
	}
	if a.Currency.Prefix {
		return a.Currency.Symbol + s
	}
	return s + a.Currency.Symbol
}

func (a *Amount) Parse(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		a.Cents = 0
		a.Currency = nil
		return nil
	}
	var currency *Currency
	var valueStr string
	for _, c := range CurrencyBySymbol {
		if c.Prefix && strings.HasPrefix(s, c.Symbol) {
			currency = c
			valueStr = s[len(c.Symbol):]
			break
		}
		if !c.Prefix && strings.HasSuffix(s, c.Symbol) {
			currency = c
			valueStr = s[0 : len(s)-len(c.Symbol)]
			break
		}
	}
	if valueStr == "" && currency != nil {
		return errors.Errorf("invalid amount \"%s\" with currency symbol and no value", s)
	}
	//value Str must now be XXX for full currency or X.XX for cents
	l := len(valueStr)
	if l >= 4 && valueStr[l-2] == '.' {
		//with cents
		if i64, err := strconv.ParseInt(valueStr[0:l-3]+valueStr[l-2:], 10, 64); err != nil {
			return errors.Errorf("invalid amount \"%s\" expected X.XX", s)
		} else {
			a.Cents = int(i64)
		}
	} else {
		//full currency, no cents
		if i64, err := strconv.ParseInt(valueStr, 10, 64); err != nil {
			return errors.Errorf("invalid amount \"%s\" expected integer value", s)
		} else {
			a.Cents = int(i64) * 100
		}
	}
	return nil
}

type Currency struct {
	Name   string //e.g. "South African Rand"
	Symbol string //e.g. "R"
	Prefix bool   //true then Rxxx, false then xxx$
}

var (
	DefaultCurrency  *Currency
	CurrencyBySymbol = map[string]*Currency{}
)

func AddCurrency(c Currency) {
	if _, ok := CurrencyBySymbol[c.Symbol]; ok {
		panic(fmt.Sprintf("duplicate currency: %+v", c))
	}
	CurrencyBySymbol[c.Symbol] = &c
	if DefaultCurrency == nil {
		DefaultCurrency = &c
	}
}

func init() {
	AddCurrency(Currency{
		Name:   "South African Rand",
		Symbol: "R",
		Prefix: true,
	})
}
