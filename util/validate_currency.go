package util

var SupportedCurrencies = []string{
	"USD",
	"EUR",
	"CAD",
}

var currencySet map[string]bool

func init(){
	currencySet = make(map[string]bool, len(SupportedCurrencies))

	for _, c := range SupportedCurrencies{
		currencySet[c] = true
	}
}



func IsSupportedCurrency(currency string) bool {
	return currencySet[currency]
}