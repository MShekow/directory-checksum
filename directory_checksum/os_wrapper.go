package directory_checksum

import "os"

// OsWrapper is a simple (incomplete) wrapper for the "os" standard library package, only wrapping selected functions.
// It was created to allow unit testing code to mock the results returned by these functions.
type OsWrapper interface {
	Getwd() (dir string, err error)
}

// OsWrapperNative implements all methods defined by OsWrapper, forwarding the calls to the "os" package.
type OsWrapperNative struct{}

func (osWrapper OsWrapperNative) Getwd() (dir string, err error) {
	dir, err = os.Getwd()
	return
}
