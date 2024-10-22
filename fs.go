package apifs

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

type fs struct {
	api API
}

func (f *fs) Root() (fusefs.Node, error) {
	return dir{f.api}, nil
}

type dir struct {
	api API
}

func (d dir) Attr(ctx context.Context, a *fuse.Attr) error {
	// Use 0 for dynamic Inode id
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d dir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	return newFile(ctx, name, d.api)
}

func (d dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	items := d.api.List(ctx)
	var list []fuse.Dirent
	for _, item := range items {
		list = append(list,
			fuse.Dirent{
				Inode: item.Id,
				Name:  item.Name,
				Type:  fuse.DT_File,
			},
		)
	}
	return list, nil
}

// HandleReader
// HandleReadDirer
// HandleWriter
// HandleFlusher
//
// no state - always call the api
type file struct {
	Id   uint64
	Name string
	api  API
}

// new file from API lookup
func newFile(ctx context.Context, name string, api API) (file, error) {
	var f file
	item, err := api.Lookup(ctx, name)
	if err != nil {
		return f, fmt.Errorf("%w: %v", syscall.ENOENT, err)
	}
	f.Id = item.Id
	f.Name = item.Name
	f.api = api
	return f, nil
}

func (f file) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.Id
	a.Mode = 0600
	item, err := f.api.Lookup(ctx, f.Name)
	if err != nil {
		return fmt.Errorf("%w: %v", syscall.ENOENT, err)
	}
	a.Size = uint64(len(item.Data))
	return nil
}

func (f file) ReadAll(ctx context.Context) ([]byte, error) {
	item, err := f.api.Lookup(ctx, f.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", syscall.ENOENT, err)
	}
	return item.Data, nil
}

func (f file) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fusefs.Handle, error) {
	return newFile(ctx, f.Name, f.api)
}

func (f file) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return nil
}

func (f file) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	item, err := f.api.Lookup(ctx, f.Name)
	if err != nil {
		return fmt.Errorf("%w: %v", syscall.ENOENT, err)
	}
	resp.Data = []byte(item.Data[req.Offset : req.Offset+int64(req.Size)])
	return nil
}

func (f file) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Printf("writing %s to %v", req.Data, f)
	defer func() {
		resp.Size = len(req.Data)
	}()
	if len(req.Data) == 0 {
		// this is not a truncate call
		// but might be an error
		return nil
	}
	item, err := f.api.Lookup(ctx, f.Name)
	if err != nil {
		return fmt.Errorf("%w: %v", syscall.ENOENT, err)
	}
	// written data is longer or equal to file data at offset
	if len(req.Data) >= len(item.Data[req.Offset:]) {
		item.Data = append(item.Data[:req.Offset], req.Data...)
	} else {
		item.Data = append(
			item.Data[:req.Offset],
			append(
				req.Data,
				item.Data[req.Offset+1:]...,
			)...,
		)
	}
	return f.api.Write(ctx, f.Name, item.Data)
}
