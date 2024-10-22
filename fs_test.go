package apifs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/exp/maps"

	"bazil.org/fuse"
)

var ErrNotFound = errors.New("not found")

// in-mem api implementation
type inMemAPI struct {
	Items map[string]Item
}

func (api *inMemAPI) List(ctx context.Context) []Item {
	return maps.Values(api.Items)
}

func (api *inMemAPI) Lookup(ctx context.Context, name string) (Item, error) {
	item, ok := api.Items[name]
	if !ok {
		return Item{}, fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	return item, nil
}

func (api *inMemAPI) Write(ctx context.Context, name string, b []byte) error {
	item, ok := api.Items[name]
	if !ok {
		return fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	item.Data = b
	api.Items[name] = item
	return nil
}

func TestFileOps(t *testing.T) {
	item := Item{
		Id:   2,
		Name: "foo",
		Data: []byte("data"),
	}
	api := &inMemAPI{
		Items: map[string]Item{
			"foo": item,
		},
	}
	ctx := context.Background()
	f, err := newFile(ctx, "foo", api)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Attr", func(t *testing.T) {
		var attr fuse.Attr
		err := f.Attr(ctx, &attr)
		if err != nil {
			t.Fatal(err)
		}
		if attr.Size != uint64(len(item.Data)) {
			t.Fatal("attr.Size failure")
		}
	})
	t.Run("ReadAll", func(t *testing.T) {
		b, err := f.ReadAll(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, item.Data) {
			t.Fatal("ReadAll failure")
		}
	})
	t.Run("Read", func(t *testing.T) {
		req := &fuse.ReadRequest{
			Offset: int64(1),
			Size:   3,
		}
		res := &fuse.ReadResponse{}
		err := f.Read(ctx, req, res)
		if err != nil {
			t.Fatal(err)
		}
		got := res.Data
		want := []byte("ata")
		if !bytes.Equal(got, want) {
			t.Fatalf("got %s, want %s", got, want)
		}
	})
	// t.Run("Write", func(t *testing.T) {
	// 	tests := []struct {
	// 		name     string
	// 		original string
	// 		offset   int64
	// 		data     string
	// 		want     string
	// 	}{
	// 		{"noop", "data", 0, "", "data"},
	// 		{"replace same len", "data", 0, "text", "text"},
	// 		{"replace longer", "data", 0, "longertext", "longertext"},
	// 		{"replace and keep", "data", 1, "u", "duta"},
	// 		{"append and grow", "data", 2, "text", "datext"},
	// 	}
	// 	for _, test := range tests {
	// 		t.Run(test.name, func(t *testing.T) {
	// 			f.Data = []byte(test.original)
	// 			req := &fuse.WriteRequest{
	// 				Offset: test.offset,
	// 				Data:   []byte(test.data),
	// 			}
	// 			res := &fuse.WriteResponse{}
	// 			err := f.Write(ctx, req, res)
	// 			if err != nil {
	// 				t.Fatal(err)
	// 			}
	// 			got := string(f.Data)
	// 			want := test.want
	// 			if got != want {
	// 				t.Fatalf("got %s, want %s", got, want)
	// 			}
	// 		})
	// 	}
	// })
}
