package apifs

import (
	"context"
	"errors"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

var ErrNilAPI = errors.New("api cannot be nil")

type Item struct {
	// Must be unique
	Id   uint64
	Name string
	Data []byte
}

type API interface {
	List(context.Context) []Item
	Lookup(context.Context, string) (Item, error)
	// Read(context.Context, string) []byte
	Write(context.Context, string, []byte) error
	// Create(context.Context, Item) error
	// Delete(context.Context, string) error
}

func Mount(api API, mountPoint string) error {
	if api == nil {
		return ErrNilAPI
	}
	c, err := fuse.Mount(
		mountPoint,
		fuse.FSName("apifs"),
		fuse.Subtype("apifs"),
	)
	if err != nil {
		return err
	}
	defer c.Close()
	defer fuse.Unmount(mountPoint)
	f := &fs{api}
	return fusefs.Serve(c, f)
}
