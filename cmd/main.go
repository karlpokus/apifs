package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"golang.org/x/exp/maps"

	"github.com/karlpokus/apifs"
)

var ErrNotFound = errors.New("not found")

// in-mem api implementation
type API struct {
	Items map[string]apifs.Item
}

func (api *API) List(ctx context.Context) []apifs.Item {
	return maps.Values(api.Items)
}

func (api *API) Lookup(ctx context.Context, name string) (apifs.Item, error) {
	item, ok := api.Items[name]
	if !ok {
		return apifs.Item{}, fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	return item, nil
}

func (api *API) Write(ctx context.Context, name string, b []byte) error {
	item, ok := api.Items[name]
	if !ok {
		return fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	item.Data = b
	api.Items[name] = item
	return nil
}

func main() {
	api := &API{
		Items: map[string]apifs.Item{
			"foo": {
				Id:   2,
				Name: "foo",
				Data: []byte("bar"),
			},
			"moo": {
				Id:   3,
				Name: "moo",
				Data: []byte("cow"),
			},
		},
	}
	err := apifs.Mount(api, "/tmp/mnt")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}
