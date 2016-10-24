package airlift

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	r "github.com/dancannon/gorethink"
)

// ErrNotFound is the error that is returned if the document being queried
// could not be found.
var ErrNotFound = errors.New("airlift: not found")

const enableProfiling = false

var session *r.Session

// Connect connects to the database
func Connect(opts r.ConnectOpts) error {
	var err error
	session, err = r.Connect(opts)

	return err
}

func getOne(term r.Term, result interface{}) error {
	start := time.Now()
	c, err := term.Run(session, r.RunOpts{
		Profile: enableProfiling,
	})
	if err != nil {
		return err
	}

	if enableProfiling {
		if time.Since(start) > time.Millisecond*30 {
			fmt.Println("Slow query warning, took", time.Since(start).Seconds(), "seconds")
			resp, _ := json.MarshalIndent(c.Profile(), "", "  ")
			fmt.Println(string(resp))
		}
	}

	return c.One(result)
}

func getAll(term r.Term, result interface{}) error {
	start := time.Now()
	c, err := term.Run(session, r.RunOpts{
		Profile: enableProfiling,
	})
	if err != nil {
		return err
	}

	if enableProfiling {
		if time.Since(start) > time.Millisecond*30 {
			fmt.Println("Slow query warning, took", time.Since(start).Seconds(), "seconds")
			resp, _ := json.MarshalIndent(c.Profile(), "", "  ")
			fmt.Println(string(resp))
		}
	}

	return c.All(result)
}
