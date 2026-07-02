package util

import (
	"fmt"
	"math/rand"
	"strings"
)

func RandomInt(min, max int64) int64 {
	return rand.Int63n(max-min+1) + min
}

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	var sb strings.Builder
	k := len(letters)
	for range n {
		c := letters[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomUser() string {
	return RandomString(6)
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
