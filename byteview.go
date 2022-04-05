package simple_groupcache

type ByteView struct {
	data []byte
}

func (v ByteView) Len() int {
	return len(v.data)
}

func (v ByteView) ByteSlice() []byte {
	cp := make([]byte, v.Len())
	copy(cp, v.data)
	return cp
}
