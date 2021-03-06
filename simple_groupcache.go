package simple_groupcache

import (
	"context"
	"log"
	"simple_groupcache/lru"
	"simple_groupcache/pb"
	"simple_groupcache/singleflight"
	"sync"
)

var (
	mu       = sync.RWMutex{}
	groupMap = make(map[string]*Group)
)

type Getter interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

type GetterFunc func(ctx context.Context, key string) ([]byte, error)

func (f GetterFunc) Get(ctx context.Context, key string) ([]byte, error) {
	return f(ctx, key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name   string
	getter Getter
	// 节点选择接口
	peers  PeerPicker
	loader *singleflight.Group

	mainCache *cache
	// groupcache还提供了热点数据多点缓存的功能
	// hotCache *cache
}

func NewGroup(name string, maxEntries int, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := groupMap[name]; ok {
		panic("duplicate registration of group " + name)
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: newCache(maxEntries),
		loader:    &singleflight.Group{},
	}
	groupMap[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groupMap[name]
}

func (g *Group) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := g.mainCache.get(key); ok {
		return val.ByteSlice(), nil
	}
	val, err := g.load(ctx, key)
	if err != nil {
		return nil, err
	}
	return val.ByteSlice(), nil
}

// 缓存未命中 从其它地方获取数据
func (g *Group) load(ctx context.Context, key string) (ByteView, error) {
	val, err := g.loader.Do(key, func() (interface{}, error) {
		p, ok := g.peers.PickPeer(key)
		if !ok {
			bv, err := g.getLocally(ctx, key)
			if err != nil {
				return ByteView{}, err
			}
			g.populateCache(key, bv)
			return bv, err
		}
		return g.getPeer(ctx, key, p)
	})
	if err != nil {
		return ByteView{}, err
	}
	return val.(ByteView), nil
}

// 直接从数据源获取数据
func (g *Group) getLocally(ctx context.Context, key string) (ByteView, error) {
	log.Println("getLocally:" + key)
	val, err := g.getter.Get(ctx, key)
	if err != nil {
		return ByteView{}, err
	}
	bv := ByteView{data: val}
	g.populateCache(key, bv)
	return bv, nil
}

// 从对等点处获取数据
func (g *Group) getPeer(ctx context.Context, key string, getter PeerGetter) (ByteView, error) {
	log.Println("getPeer:" + g.name + ":" + key)
	req := &pb.GetRequest{
		Group: &g.name,
		Key:   &key,
	}
	res := &pb.GetResponse{}
	err := getter.Get(ctx, req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{data: res.Value}, nil
}

// SetPeerPicker 将对等点选择接口注入到group中
func (g *Group) SetPeerPicker(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 缓存数据
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) Name() string {
	return g.name
}

func newCache(maxEntries int) *cache {
	return &cache{
		mu:         sync.Mutex{},
		lru:        nil,
		maxEntries: maxEntries,
	}
}

// 对lru进行封装使其支持并发读写
type cache struct {
	mu  sync.Mutex
	lru *lru.Cache
	// 直接用键值对数量来限制缓存大小，实际应该统计key value的内存占用
	maxEntries int
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化
	if c.lru == nil {
		c.lru = lru.New(c.maxEntries)
	}
	c.lru.Put(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	v, ok := c.lru.Get(key)
	if !ok {
		return
	}
	return v.(ByteView), ok
}
