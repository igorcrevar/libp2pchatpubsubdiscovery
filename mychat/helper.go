package mychat

import (
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var isInitialized bool = false

func GenerateRandString(n int) string {
	if !isInitialized {
		rand.Seed(time.Now().UnixNano())
		isInitialized = true
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
