package simple_groupcache

import (
	"context"
	"errors"
	"log"
	"net/http"
	"testing"
)

var data = map[string]string{
	"a": "aa",
	"b": "bb",
	"1": "11111111111111111111111111111111111",
}

func TestGroupCache(t *testing.T) {
	g := NewGroup("test", 100, GetterFunc(func(ctx context.Context, key string) ([]byte, error) {
		if val, ok := data[key]; ok {
			return []byte(val), nil
		}
		return nil, errors.New("not find")
	}))
	prefix := "http://"
	self := "127.0.0.1:8001"
	peers := []string{
		"http://127.0.0.1:8001",
		"http://127.0.0.1:8002",
	}
	p := NewHTTPPool(prefix+self, "")
	g.SetPeerPicker(p)
	p.Set(peers...)
	log.Fatal(http.ListenAndServe(self, p))
}

func TestGroupCache2(t *testing.T) {
	g := NewGroup("test", 100, GetterFunc(func(ctx context.Context, key string) ([]byte, error) {
		if val, ok := data[key]; ok {
			return []byte(val), nil
		}
		return nil, errors.New("not find")
	}))
	prefix := "http://"
	self := "127.0.0.1:8002"
	peers := []string{
		"http://127.0.0.1:8001",
		"http://127.0.0.1:8002",
	}
	p := NewHTTPPool(prefix+self, "")
	g.SetPeerPicker(p)
	p.Set(peers...)
	log.Fatal(http.ListenAndServe(self, p))
}
