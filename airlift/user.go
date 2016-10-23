package airlift

import (
	"encoding/json"
	"fmt"
	"time"

	r "github.com/dancannon/gorethink"
)

// User represents a user.
type User struct {
	Name        string   `gorethink:"name"`
	Username    string   `gorethink:"username"`
	Email       string   `gorethink:"email"`
	Password    string   `gorethink:"password,omitempty"`
	Schedule    []string `gorethink:"schedule,omitempty"`
	RawSchedule []string `gorethink:"raw_schedule,omitempty"`
}

func getOne(term r.Term, result interface{}) error {
	start := time.Now()
	c, err := term.Run(session, r.RunOpts{
		Profile: true,
	})
	if err != nil {
		return err
	}

	if time.Since(start) > time.Millisecond*90 {
		fmt.Println("Slow query warning, took", time.Since(start).Seconds(), "seconds")
		resp, _ := json.MarshalIndent(c.Profile(), "", "  ")
		fmt.Println(string(resp))
	}

	return c.One(result)
}

func getAll(term r.Term, result interface{}) error {
	start := time.Now()
	c, err := term.Run(session, r.RunOpts{
		Profile: true,
	})
	if err != nil {
		return err
	}

	if time.Since(start) > time.Millisecond*90 {
		fmt.Println("Slow query warning, took", time.Since(start).Seconds(), "seconds")
		resp, _ := json.MarshalIndent(c.Profile(), "", "  ")
		fmt.Println(string(resp))
	}

	return c.All(result)
}

// GetUser returns the user data of a user.
func GetUser(username string) (User, error) {
	user := User{}
	err := getOne(r.Table("users").Get(username).Default(User{}), &user)
	if err != nil {
		return User{}, err
	}
	if user.Username == "" {
		return User{}, ErrNotFound
	}

	return user, nil
}

// UpdatePassword updates the password of a user with the given hash.
func UpdatePassword(username, hash string) error {
	_, err := r.Table("users").Get(username).Update(struct {
		Password string `gorethink:"password"`
	}{hash}).RunWrite(session)
	if err != nil {
		return err
	}

	return nil
}
