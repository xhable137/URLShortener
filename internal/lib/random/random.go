package random

import (
	"math/rand"
	"time"
)

func NewRandomString(size int) string {
	// Use a combination of time and a random seed for better uniqueness
	seed := time.Now().UnixNano() + rand.Int63n(1000000)
	rnd := rand.New(rand.NewSource(seed))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" +
		"abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}
	return string(b)
}
