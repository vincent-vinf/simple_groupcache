package simple_groupcache

import (
	"context"
	"errors"
	"log"
	"net/http"
	"testing"
)

func TestGroupCache(t *testing.T) {
	m := make(map[string]string)
	m["1"] = "1"
	m["2"] = "2"
	m["3"] = "2"

	NewGroup("test", 100, GetterFunc(func(ctx context.Context, key string) ([]byte, error) {
		if val, ok := m[key]; ok {
			return []byte(val), nil
		}
		return nil, errors.New("not find")
	}))
	p := NewHTTPPool("http://127.0.0.1:8001", "")
	log.Fatal(http.ListenAndServe("127.0.0.1:8001", p))
}
