package utils

import "testing"

func TestEncrypt(t *testing.T) {
	encrypt, err := Encrypt("LDQ20020618xxx")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(encrypt)
}
