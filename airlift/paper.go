package airlift

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	r "github.com/dancannon/gorethink"
)

// Paper represents a user uploaded paper
type Paper struct {
	ID            string    `gorethink:"id"`
	Title         string    `gorethink:"title"`
	Year          int       `gorethink:"year"`
	Subject       string    `gorethink:"subject"`
	Author        string    `gorethink:"author"`
	Public        bool      `gorethink:"public"`
	SolutionsSize uint64    `gorethink:"solutions_size,omitempty"`
	SourceSize    uint64    `gorethink:"source_size,omitempty"`
	QuestionsSize uint64    `gorethink:"questions_size,omitempty"`
	Completed     []string  `gorethink:"completed,omitempty"`
	NumCompleted  int       `gorethink:"num_completed,omitempty"`
	Uploader      string    `gorethink:"uploader"`
	UpdatedTime   time.Time `gorethink:"updated_time"`
	UploadTime    time.Time `gorethink:"upload_time"`
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

// UpdatePaper updates the information of a paper
func UpdatePaper(paper Paper) error {
	paper.UpdatedTime = time.Now()
	paper.UploadTime = time.Time{}

	result, err := r.Table("papers").Get(paper.ID).Update(paper).RunWrite(session)
	if err != nil {
		return err
	}

	if result.Errors > 0 {
		return errors.New("airlift: " + result.FirstError)
	}

	return nil
}

// DeletePaper deletes a paper
func DeletePaper(id string) error {
	result, err := r.Table("papers").Get(id).Delete().RunWrite(session)
	if err != nil {
		return err
	}

	if result.Errors > 0 {
		return errors.New("airlift: " + result.FirstError)
	}

	return nil
}

// NewPaper creates a new paper and returns a unique human friendly ID.
// Data must be populated using UpdatePaper.
func NewPaper(title string) (string, error) {
	var paper Paper

	paper.Title = title
	paper.UploadTime = time.Now()

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
