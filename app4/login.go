package app4

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const loginURL = "https://ccgswa.app4.ws/student/logincheck_student.php"
const homepageURL = "https://ccgswa.app4.ws/portal/index.php"

// ErrInvalidCredentials is the error that is returned by Login if
// invalid credentials are provided.
var ErrInvalidCredentials = errors.New("app4: invalid credentials")

// Session represents a logged in user's session.
type Session struct {
	*http.Client
}

// GetClassNames returns the raw names on the logged in user's timetable.
func (s Session) GetClassNames() ([]string, error) {
	resp, err := s.Client.Get("https://ccgswa.app4.ws/portal/timetable.php")
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	classMap := make(map[string]bool)

	subjects := doc.Find(".ttsub")
	for i := 0; i < subjects.Length(); i++ {
		className := subjects.Eq(i).Text()
		if !classMap[className] {
			classMap[className] = true
		}
	}

	var classes []string
	for uniqueClass := range classMap {
		classes = append(classes, uniqueClass)
	}

	if len(classes) == 0 {
		return nil, errors.New("app4: no classes found")
	}

	return classes, nil
}

// Login logs in to the app4 endpoint with the given username and password.
// If unsucessful a non-nil error will be returned.
func Login(username string, password string) (Session, error) {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar:     jar,
		Timeout: time.Second * 20,
	}

	resp, err := client.PostForm(loginURL, url.Values{
		"username":     {username},
		"userpassword": {password},
	})
	if err != nil {
		return Session{}, err
	}

	if resp.Request.URL.String() == loginURL {
		return Session{}, ErrInvalidCredentials
	}

	if resp.Request.URL.String() == homepageURL {
		return Session{client}, nil
	}

	return Session{}, errors.New("app4: unknown response")
}
