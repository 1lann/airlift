package airlift

import (
	"errors"

	r "github.com/dancannon/gorethink"
)

// ErrNotFound is the error that is returned if the document being queried
// could not be found.
var ErrNotFound = errors.New("airlift: not found")

var session *r.Session

// Connect connects to the database
func Connect(opts r.ConnectOpts) error {
	var err error
	session, err = r.Connect(opts)

	return err
}
