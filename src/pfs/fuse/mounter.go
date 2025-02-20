package fuse

import (
	"os"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
)

const (
	namePrefix = "pfs://"
	subtype    = "pfs"
)

type mounter struct {
	address   string
	apiClient pfs.APIClient
}

func newMounter(address string, apiClient pfs.APIClient) Mounter {
	return &mounter{
		address,
		apiClient,
	}
}

func (m *mounter) Mount(
	mountPoint string,
	shard *pfs.Shard,
	commits []*pfs.Commit,
	ready chan bool,
) (retErr error) {
	var once sync.Once
	defer once.Do(func() {
		if ready != nil {
			close(ready)
		}
	})
	// TODO: should we make the caller do this?
	if err := os.MkdirAll(mountPoint, 0777); err != nil {
		return err
	}
	name := namePrefix + m.address
	conn, err := fuse.Mount(
		mountPoint,
		fuse.FSName(name),
		fuse.VolumeName(name),
		fuse.Subtype(subtype),
		fuse.AllowOther(),
		fuse.WritebackCache(),
		fuse.MaxReadahead(1<<32-1),
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	once.Do(func() {
		if ready != nil {
			close(ready)
		}
	})
	if err := fs.Serve(conn, newFilesystem(m.apiClient, shard, commits)); err != nil {
		return err
	}
	<-conn.Ready
	return conn.MountError
}

func (m *mounter) Unmount(mountPoint string) error {
	return fuse.Unmount(mountPoint)
}
