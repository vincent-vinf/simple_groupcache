package simple_groupcache

import (
	"context"
	"simple_groupcache/pb"
)

type PeerGetter interface {
	Get(ctx context.Context, in *pb.GetRequest, out *pb.GetResponse) error
}

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
