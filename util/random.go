package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)
const alphabet = "abcdefghijklmopqrstuvwxyz"
func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer between min and max
func RandomInt(min,max int64) int64{
	return min + rand.Int63n(max - min + 1)
}

func RandomString(n int)string{
	var sb strings.Builder
	k := len(alphabet)

	for range n{
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return  sb.String()
}

func RandomOwner() string{
	return RandomString(6)
}

func RandomMoney()decimal.Decimal{
	return decimal.NewFromInt(RandomInt(0,1000))
}

func RandomCurrency()string{
	currencies := SupportedCurrencies
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

func RandomEmail()string{
	return  fmt.Sprintf("%s@gmail.com",RandomString(5))
}