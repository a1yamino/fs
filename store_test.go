package main

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	paths := []string{"a", "b", "c"}

	joined := path.Join(paths...)
	assert.Equal(t, "a/b/c", joined)
	t.Logf("path.joined: %s", joined)

	joined = strings.Join(paths, "/")
	assert.Equal(t, "a/b/c", joined)
	t.Logf("strings.joined: %s", joined)
}

func BenchmarkJoin(b *testing.B) {
	paths := []string{"a", "b", "c"}

	b.Run("path.Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path.Join(paths...)
		}
	})

	b.Run("strings.Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strings.Join(paths, "/")
		}
	})
}

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"

	if !assert.Equal(t, expectedPathName, pathKey.PathName) {
		t.Errorf("expectedPathName: %s, got: %s", expectedPathName, pathKey.PathName)
	}

	if !assert.Equal(t, expectedFilename, pathKey.Filename) {
		t.Errorf("expectedFilename: %s, got: %s", expectedFilename, pathKey.Filename)
	}

}
func TestStore(t *testing.T) {
	s := newStore()
	id := generateID()
	defer teardown(t, s)

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(id, key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		r, _, err := s.Read(id, key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(id, key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(id, key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
