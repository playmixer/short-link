package util

import (
	"math/rand"
)

// RandomString генерирует строку заданой длины.
func RandomString(n uint) string {
	var letterRunes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// BuildData возвращает значение билдера.
func BuildData(data string) string {
	if data != "" {
		return data
	}
	return "N/A"
}
