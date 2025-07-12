package xserver

type VPSID string

func (v VPSID) String() string {
	return string(v)
}

type UniqueID string

func (u UniqueID) String() string {
	return string(u)
}
