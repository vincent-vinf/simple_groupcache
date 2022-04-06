package simple_groupcache

import "context"

type PeerGetter interface {
	Get(cxt context.Context, group string, key string) ([]byte, error)
}

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
