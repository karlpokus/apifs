# apifs
Mount an API to disk

Built on FUSE. Depends on [fusermount3](https://man7.org/linux/man-pages/man1/fusermount3.1.html)

My use-case is letting a common text editor manipulate API items.

This is what it looks like atm:

````go
type API interface {
	List(context.Context) []Item
	Lookup(context.Context, string) (Item, error)

    // Read(context.Context, string) []byte

	Write(context.Context, string, []byte) error
}
````

Synchronization and caching is left to the implementation.

# todos
- [ ] create mountPoint if missing
- [x] unmount on exit
- [ ] unmount on interrupt
- [ ] lib for common APIs like 1password, s3
- [ ] file permissions
- [ ] put ops logs behind verbose flag
- [ ] honor context deadline
- [ ] api item cache
- [ ] Inode id generation
- [ ] add note on vim complaining about fsync
