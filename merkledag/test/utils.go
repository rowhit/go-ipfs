package mdutils

import (
	bsrv "github.com/ipfs/go-ipfs/blockservice"
	dag "github.com/ipfs/go-ipfs/merkledag"

	offline "gx/ipfs/QmQQoWPUwZ6rfjXK2jftyeEWEn5BxKxQNwCYjt2Dm1gUQk/go-ipfs-exchange-offline"
	blockstore "gx/ipfs/QmT7sUaxknR9QJLksVN2iowzd19uPGd1g2mDm7dKVmygdV/go-ipfs-blockstore"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	ds "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
	dssync "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore/sync"
)

// Mock returns a new thread-safe, mock DAGService.
func Mock() ipld.DAGService {
	return dag.NewDAGService(Bserv())
}

// Bserv returns a new, thread-safe, mock BlockService.
func Bserv() bsrv.BlockService {
	bstore := blockstore.NewBlockstore(dssync.MutexWrap(ds.NewMapDatastore()))
	return bsrv.New(bstore, offline.Exchange(bstore))
}
