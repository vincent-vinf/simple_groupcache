package simple_groupcache

import (
	"context"
	"log"
	"testing"
)

func TestGroup(t *testing.T) {
	m := make(map[string]string)
	m["1"] = "1"
	m["2"] = "2"
	m["3"] = "2"

	g := NewGroup("test", 100, GetterFunc(func(ctx context.Context, key string) ([]byte, error) {
		if val, ok := m[key]; ok {
			return []byte(val), nil
		}
		return nil, nil
	}))
	log.Println(g.Name())
	ctx := context.Background()
	for key, value := range m {
		val, err := g.Get(ctx, key)
		if err != nil {
			log.Fatalln(err)
		}
		if string(val) != value {
			log.Fatalln("not equal")
		}
	}

}
