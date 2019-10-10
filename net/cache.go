package net

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// DefaultCachePath ...
var DefaultCachePath = "tmp"
var cache = NewCache(DefaultCachePath)

// CacheDisable ...
var CacheDisable = false

// Cache ...
type Cache struct {
	tmp string
}

// CacheOff ...
func CacheOff() {
	CacheDisable = true
}

// HasCache ...
func HasCache() bool {
	return !CacheDisable
}

func hash(url string) string {
	sum256 := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", sum256)
}

// NewCache ...
func NewCache(tmp string) *Cache {
	if tmp == "" {
		tmp = DefaultCachePath
	}
	s, e := filepath.Abs(tmp)
	if e != nil {
		panic(e)
	}
	_ = os.MkdirAll(tmp, os.ModePerm)
	return &Cache{tmp: s}
}

// Reader ...
func (c *Cache) Reader(url string) (io.ReadCloser, error) {
	e := c.Get(url)
	if e != nil {
		return nil, e
	}
	file, e := os.Open(filepath.Join(c.tmp, hash(url)))
	if e != nil {
		return nil, e
	}
	return file, nil
}

// Get ...
func (c *Cache) Get(url string) (e error) {
	name := hash(url)
	stat, e := os.Stat(filepath.Join(c.tmp, name))
	log.With("url", url, "hash", name).Info("cache get")
	if e == nil && stat.Size() != 0 {
		return nil
	}

	if cli == nil {
		cli = http.DefaultClient
	}

	res, e := cli.Get(url)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	file, e := os.OpenFile(filepath.Join(c.tmp, name), os.O_TRUNC|os.O_CREATE|os.O_RDONLY|os.O_SYNC, os.ModePerm)
	if e != nil {
		return e
	}
	defer file.Close()
	written, e := io.Copy(file, res.Body)
	if e != nil {
		return e
	}
	//ignore written
	_ = written
	return nil
}

// MoveCache ...
func MoveCache(path, url, to string) (written int64, e error) {
	info, e := os.Stat(filepath.Join(path, hash(url)))
	if e != nil && os.IsNotExist(e) {
		return written, errors.Wrap(e, "cache get error")
	}
	if info.IsDir() {
		return written, errors.New("cache get a dir")
	}
	s, e := filepath.Abs(to)
	if e != nil {
		return written, e
	}
	dir, _ := filepath.Split(s)
	_ = os.MkdirAll(dir, os.ModePerm)
	file, e := os.Open(filepath.Join(path, hash(url)))
	if e != nil {
		return written, e
	}

	openFile, e := os.OpenFile(filepath.Join(s), os.O_TRUNC|os.O_CREATE|os.O_RDONLY|os.O_SYNC, os.ModePerm)
	if e != nil {
		return written, e
	}
	return io.Copy(openFile, file)
}
