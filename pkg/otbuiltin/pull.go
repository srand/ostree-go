package otbuiltin

import (
	"unsafe"

	glib "github.com/ostreedev/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

// NewPullOptions defines all of the options for pulling refs
//
// Note: while this is private, fields are public and part of the API.
type pullOptions struct {
	// Reject writes of content objects with modes outside of 0775 (Since: 2017.7)
	BaseUserOnlyFiles bool
	// Fetch only the commit metadata
	CommitOnly bool
	// Write out refs suitable for mirrors and fetch all refs if none requested
	Mirror bool
	// Don't verify checksums of objects HTTP repositories (Since: 2017.12)
	TrustedHttp bool
	// Do verify checksums of local (filesystem-accessible) repositories (defaults on for HTTP)
	Untrusted bool
}

// NewPullOptions instantiates and returns a pullOptions struct with default values set
func NewPullOptions() pullOptions {
	return pullOptions{}
}

// Pull pulls refs from the named remote.
// Returns an error if the refs could not be fetched.
func (repo *Repo) Pull(remote string, refs []string, opts pullOptions) error {
	var cancellable *glib.GCancellable

	cremote := C.CString(remote)
	defer C.free(unsafe.Pointer(cremote))

	crefs := (**C.char)(C.malloc(C.size_t(len(refs)+1) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	defer C.free(unsafe.Pointer(crefs))

	crefsarray := (*[1 << 30]*C.char)(unsafe.Pointer(crefs))
	for idx, ref := range refs {
		crefsarray[idx] = C.CString(ref)
	}
	crefsarray[len(refs)] = (*C.char)(unsafe.Pointer(nil))

	var gerr = glib.NewGError()
	cerr := (*C.GError)(gerr.Ptr())
	defer C.free(unsafe.Pointer(cerr))

	// Process options into bitflags
	repoPullOptions := C.OstreeRepoPullFlags(C.OSTREE_REPO_PULL_FLAGS_NONE)
	if opts.BaseUserOnlyFiles != false {
		repoPullOptions |= C.OSTREE_REPO_PULL_FLAGS_BAREUSERONLY_FILES
	}
	if opts.CommitOnly {
		repoPullOptions |= C.OSTREE_REPO_PULL_FLAGS_COMMIT_ONLY
	}
	if opts.Mirror {
		repoPullOptions |= C.OSTREE_REPO_PULL_FLAGS_MIRROR
	}
	if opts.TrustedHttp {
		repoPullOptions |= C.OSTREE_REPO_PULL_FLAGS_UNTRUSTED
	}
	if opts.Untrusted {
		repoPullOptions |= C.OSTREE_REPO_PULL_FLAGS_TRUSTED_HTTP
	}

	// Pull refs from remote
	if !glib.GoBool(glib.GBoolean(C.ostree_repo_pull(repo.native(), cremote, crefs, repoPullOptions, nil, (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
		return generateError(cerr)
	}

	return nil
}
