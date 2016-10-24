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
	Completed     []string  `gorethink:"completed"`
	Uploader      string    `gorethink:"uploader"`
	SolutionsSize uint64    `gorethink:"solutions_size,omitempty"`
	SourceSize    uint64    `gorethink:"source_size,omitempty"`
	QuestionsSize uint64    `gorethink:"questions_size,omitempty"`
	HasCompleted  bool      `gorethink:"has_completed,omitempty"`
	NumCompleted  int       `gorethink:"num_completed,omitempty"`
	UpdatedTime   time.Time `gorethink:"updated_time,omitempty"`
	UploadTime    time.Time `gorethink:"upload_time,omitempty"`
}

type updatePaper struct {
	ID            string    `gorethink:"id"`
	Title         string    `gorethink:"title"`
	Year          int       `gorethink:"year"`
	Subject       string    `gorethink:"subject"`
	Author        string    `gorethink:"author"`
	Public        bool      `gorethink:"public"`
	Uploader      string    `gorethink:"uploader"`
	Completed     []string  `gorethink:"completed,omitempty"`
	SolutionsSize uint64    `gorethink:"solutions_size,omitempty"`
	SourceSize    uint64    `gorethink:"source_size,omitempty"`
	QuestionsSize uint64    `gorethink:"questions_size,omitempty"`
	UpdatedTime   time.Time `gorethink:"updated_time,omitempty"`
}

// FullPaper represents a user uploaded paper with additional data
type FullPaper struct {
	Paper
	UploaderName string `gorethink:"uploader_name"`
	SubjectName  string `gorethink:"subject_name"`
}

var rowFullPaperTitle = func(row r.Term) interface{} {
	return row.Field("year").CoerceTo("string").
		Add(" ", row.Field("subject"), " ", row.Field("title"))
}

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
	title = strconv.Itoa(paper.Year) + "_" + paper.Subject + "_" + title
	nonce := make([]byte, 2)
	_, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}

	return title + "_" + hex.EncodeToString(nonce), nil
}

// UpdatePaper updates the information of a paper
func UpdatePaper(paper Paper) error {
	updatedPaper := updatePaper{
		ID:            paper.ID,
		Title:         paper.Title,
		Year:          paper.Year,
		Subject:       paper.Subject,
		Author:        paper.Author,
		Uploader:      paper.Uploader,
		Public:        paper.Public,
		Completed:     paper.Completed,
		SolutionsSize: paper.SolutionsSize,
		SourceSize:    paper.SourceSize,
		QuestionsSize: paper.QuestionsSize,
		UpdatedTime:   time.Now(),
	}

	result, err := r.Table("papers").Get(paper.ID).Update(updatedPaper).RunWrite(session)
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
func NewPaper(title, subject string, year int) (string, error) {
	var paper Paper

	paper.Title = title
	paper.UploadTime = time.Now()
	paper.Subject = subject
	paper.Year = year

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

// SetPaperCompleted marks a practice paper as completed for a user.
func SetPaperCompleted(id, username string, completed bool) error {
	var query r.Term
	if completed {
		query = r.Table("papers").Get(id).Update(map[string]interface{}{
			"completed": r.Row.Field("completed").SetInsert(username),
		})
	} else {
		query = r.Table("papers").Get(id).Update(map[string]interface{}{
			"completed": r.Row.Field("completed").SetDifference([]string{username}),
		})
	}

	result, err := query.RunWrite(session)
	if err != nil {
		return err
	}

	if result.Errors > 0 {
		return errors.New("airlift: " + result.FirstError)
	}

	return nil
}

// GetCompletedPapers returns the papers completed by a user.
func GetCompletedPapers(username string) ([]FullPaper, error) {
	var papers []FullPaper
	err := getAll(r.Table("papers").
		GetAllByIndex("completed", username).
		OrderBy(r.Desc(rowFullPaperTitle)).
		EqJoin("uploader", r.Table("users"), r.EqJoinOpts{Index: "username"}).
		EqJoin(r.Row.Field("left").Field("subject"), r.Table("subjects"), r.EqJoinOpts{Index: "id"}).
		Map(func(row r.Term) interface{} {
			return row.Field("left").Field("left").Merge(map[string]interface{}{
				"has_completed": row.Field("left").Field("left").Field("completed").Contains(username),
				"uploader_name": row.Field("left").Field("right").Field("name"),
				"subject_name":  row.Field("right").Field("name"),
			})
		}),
		&papers)
	return papers, err
}

// GetUploadedPapers returns the papers uploaded by a user.
func GetUploadedPapers(username string) ([]FullPaper, error) {
	var papers []FullPaper
	err := getAll(r.Table("papers").
		GetAllByIndex("uploader", username).
		OrderBy(r.Desc(rowFullPaperTitle)).
		EqJoin("uploader", r.Table("users"), r.EqJoinOpts{Index: "username"}).
		EqJoin(r.Row.Field("left").Field("subject"), r.Table("subjects"), r.EqJoinOpts{Index: "id"}).
		Map(func(row r.Term) interface{} {
			return row.Field("left").Field("left").Merge(map[string]interface{}{
				"has_completed": row.Field("left").Field("left").Field("completed").Contains(username),
				"uploader_name": row.Field("left").Field("right").Field("name"),
				"subject_name":  row.Field("right").Field("name"),
			})
		}),
		&papers)
	return papers, err
}

// GetFullPaper returns the paper with additional data uploaded by a user.
func GetFullPaper(id, username string) (FullPaper, error) {
	var paper FullPaper
	err := getOne(r.Table("papers").Get(id).
		Merge(func(row r.Term) interface{} {
			return map[string]interface{}{
				"has_completed": row.Field("completed").Contains(username),
				"uploader_name": r.Table("users").Get(row.Field("uploader")).
					Field("name").Default("Unknown"),
				"subject_name": r.Table("subjects").Get(row.Field("subject")).
					Field("name").Default("Unknown"),
			}
		}), &paper)
	return paper, err
}
