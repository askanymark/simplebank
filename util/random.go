package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomInt returns a random integer between the given min and max values inclusively.
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of the specified length composed of lowercase English alphabet characters.
func RandomString(length int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < length; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generates a random string of 6 lowercase alphabetic characters to represent an owner's name.
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random monetary value between 0 and 1000 inclusively and returns it as an int64.
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency returns one of the supported currencies at random
func RandomCurrency() string {
	currencies := []string{EUR, USD, GBP}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

// RandomEmail returns an email address using a random 6-character string followed by "@email.com".
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
