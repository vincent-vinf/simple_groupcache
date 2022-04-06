package simple_groupcache

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"simple_groupcache/consistenthash"
	"simple_groupcache/pb"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/cache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	basePath string
	self     string
	// 节点选择功能
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string, path string) *HTTPPool {
	p := &HTTPPool{
		self: self,
		mu:   sync.Mutex{},
	}
	if path != "" {
		p.basePath = path
	} else {
		p.basePath = defaultBasePath
	}
	http.Handle(self, p)
	return p
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]
	// 获取group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	ctx := r.Context()
	log.Println("key:" + key)
	data, err := group.Get(ctx, key)
	log.Println("data:" + string(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-pb")
	body, err := proto.Marshal(&pb.GetResponse{Value: data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(body)
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	u := p.peers.Get(key)
	log.Println(u)
	if u == p.self {
		return nil, false
	}
	return p.httpGetters[u], true
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas)
	p.peers.AddNode(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseUrl: peer + p.basePath}
	}
}

// http客户端
type httpGetter struct {
	baseUrl string
}

func (g *httpGetter) Get(ctx context.Context, in *pb.GetRequest, out *pb.GetResponse) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		g.baseUrl,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	log.Println(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	err = proto.Unmarshal(bytes, out)
	if err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}
