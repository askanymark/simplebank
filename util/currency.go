package util

const (
	USD = "USD"
	EUR = "EUR"
	GBP = "GBP"
)

// IsSupportedCurrency returns false if provided currency is unsupported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, GBP:
		return true
	}

	return false
}
