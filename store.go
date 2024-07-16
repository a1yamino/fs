package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const defaultRootFolderName = "fs"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		// use path.Join for cross-platform compatibility
		PathName: path.Join(paths...),
		Filename: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, string(filepath.Separator))
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {
	return path.Join(p.PathName, p.Filename)
}

type StoreOpts struct {
	// Root is the root folder for the store
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opt StoreOpts) *Store {
	if opt.PathTransformFunc == nil {
		opt.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opt.Root) == 0 {
		opt.Root = defaultRootFolderName
	}
	return &Store{StoreOpts: opt}
}

func (s *Store) Has(id string, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPath := path.Join(s.Root, id, pathKey.FullPath())

	_, err := os.Stat(fullPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(id string, key string) error {
	pathKey := s.PathTransformFunc(key)

	defer log.Printf("delete [%s] from disk\n", pathKey.Filename)
	fullPath := path.Join(s.Root, id, pathKey.FullPath())

	// TODO: empty folder
	return os.Remove(fullPath)
}

func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	fullPath := path.Join(s.Root, id, pathKey.PathName)

	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return nil, err
	}

	fullPath = path.Join(fullPath, pathKey.Filename)

	return os.Create(fullPath)
}

func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Store) Read(id string, key string) (io.Reader, int64, error) {
	return s.readStream(id, key)
}

func (s *Store) readStream(id string, key string) (io.Reader, int64, error) {
	pathKey := s.PathTransformFunc(key)
	fullPath := path.Join(s.Root, id, pathKey.FullPath())

	f, err := os.Open(fullPath)
	if err != nil {
		return nil, 0, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	return f, fi.Size(), nil
}
