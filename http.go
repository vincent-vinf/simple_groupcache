package simple_groupcache

import (
	"net/http"
	"strings"
)

type HTTPPool struct {
	basePath string
	self     string
}

const defaultBasePath = "/cache/"

func NewHTTPPool(self string, path string) *HTTPPool {
	p := &HTTPPool{
		self: self,
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
	data, err := group.Get(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//w.Header().Set("Content-Type", "application/x-protobuf")
	// 设置http请求标头为8位字节流
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// http客户端
type httpGetter struct {
}
