package airlift

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	r "github.com/dancannon/gorethink"
)

// Paper represents a user uploaded paper
type Paper struct {
	ID           string    `gorethink:"id"`
	Title        string    `gorethink:"title"`
	Year         int       `gorethink:"year"`
	Public       bool      `gorethink:"public"`
	Subject      string    `gorethink:"subject"`
	Author       string    `gorethink:"author"`
	HasSolutions bool      `gorethink:"has_solutions"`
	HasSource    bool      `gorethink:"has_source"`
	Completed    []string  `gorethink:"completed"`
	NumCompleted int       `gorethink:"num_completed,omitempty"`
	Uploader     string    `gorethink:"uploader"`
	UpdatedTime  time.Time `gorethink:"updated_time"`
	UploadTime   time.Time `gorethink:"upload_time"`
}

// TODO: Make a get full paper and use it in download.go

// GetPaper returns a paper given its ID
func GetPaper(id string) (Paper, error) {
	var paper Paper
	err := getOne(r.Table("papers").Get(id).Default(map[string]string{}), &paper)
	if err != nil {
		return Paper{}, err
	}

	return paper, nil
}

func idFromPaper(paper Paper) (string, error) {
	title := strings.ToLower(strings.Replace(paper.Title, " ", "_", -1))
	title = strconv.Itoa(paper.Year) + "_" + title
	nonce := make([]byte, 2)
	_, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}

	return title + "_" + hex.EncodeToString(nonce), nil
}

// NewPaper creates a new paper and returns a unique human friendly ID.
// The title must be filesystem safe.
func NewPaper(paper Paper) (string, error) {
	paper.UploadTime = time.Now()
	paper.UpdatedTime = paper.UploadTime

	for {
		var err error
		paper.ID, err = idFromPaper(paper)
		if err != nil {
			return "", err
		}

		result, err := r.Table("papers").Insert(paper).RunWrite(session)
		if err != nil {
			return "", err
		}

		if result.Errors == 0 {
			break
		}
	}

	return paper.ID, nil
}
