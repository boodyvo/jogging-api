package lib

import (
	"fmt"
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func CreateEmail() string {
	return fmt.Sprintf("%s@gmail.com", randString(10))
}

func CreateName() string {
	return fmt.Sprintf("%s", randString(10))
}

func CreateLocation() Location {
	return Location{
		// from 30 to 50 for testing
		Longitude: (float64(rand.Int31n(2000000)) + 3000000) / 100000,
		// from 40 to 60 for testing
		Latitude: (float64(rand.Int31n(2000000)) + 4000000) / 100000,
	}
}
