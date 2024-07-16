package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenerateID(t *testing.T) {
	got := generateID()
	if got == "" {
		t.Errorf("generateID() = %v; want a non-empty string", got)
	}
	t.Logf("Generated ID: %v, len: %d", got, len(got))
}

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(len(payload))
	fmt.Println(len(dst.String()))

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != payload {
		t.Errorf("decryption failed!!!")
	}
}
