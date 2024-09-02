// Модуль rest предоставляет http сервер и методы взаимодействия с REST API.
package rest

import (
	"bytes"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestSecretKey(t *testing.T) {
	s := &Server{}
	secret := []byte("test")
	SecretKey(secret)(s)
	if !bytes.Equal(s.secretKey, secret) {
		t.Fatal("secret key not set")
	}
}

func TestLogger(t *testing.T) {
	s := &Server{}
	lgr := zap.NewNop()
	Logger(lgr)(s)
	if reflect.ValueOf(lgr).Pointer() != reflect.ValueOf(s.log).Pointer() {
		t.Fatal("logger not set")
	}
}
