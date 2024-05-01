package util

import (
	"fmt"
	"math/rand"
)

func RandomString(len uint) string {
	return fmt.Sprintf("%x", rand.Uint32())
}
