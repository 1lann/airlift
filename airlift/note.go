package airlift

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	r "github.com/dancannon/gorethink"
)

// Note represents a user uploaded note file
type Note struct {
	ID          string    `gorethink:"id"`
	Title       string    `gorethink:"title"`
	Subject     string    `gorethink:"subject"`
	Author      string    `gorethink:"author"`
	Uploader    string    `gorethink:"uploader"`
	Public      bool      `gorethink:"public"`
	Size        uint64    `gorethink:"size,omitempty"`
	Stars       []string  `gorethink:"stars,omitempty"`
	NumStars    int       `gorethink:"num_stars,omitempty"`
	UpdatedTime time.Time `gorethink:"updated_time,omitempty"`
	UploadTime  time.Time `gorethink:"upload_time,omitempty"`
}

// FullNote represents a note with some additional data
type FullNote struct {
	Note
	UploaderName string `gorethink:"uploader_name"`
	SubjectName  string `gorethink:"subject_name"`
}

// GetFullNote returns a note and some additional data given its ID
func GetFullNote(id string) (FullNote, error) {
	var note FullNote
	err := getOne(r.Table("notes").Get(id).Merge(func(row r.Term) interface{} {
		return map[string]interface{}{
			"uploader_name": r.Table("users").Get(row.Field("uploader")).Field("name"),
			"subject_name":  r.Table("subjects").Get(row.Field("subject")).Field("name"),
			"num_stars":     row.Field("stars").Count(),
		}
	}).Default(map[string]string{}), &note)
	if err != nil {
		return FullNote{}, err
	}

	return note, nil
}

// GetNote returns a note given its ID
func GetNote(id string) (Note, error) {
	var note Note
	err := getOne(r.Table("notes").Get(id), &note)
	if err != nil {
		return Note{}, err
	}

	return note, nil
}

func idFromNote(note Note) (string, error) {
	title := strings.ToLower(strings.Replace(note.Title, " ", "_", -1))
	nonce := make([]byte, 2)
	_, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}

	return title + "_" + hex.EncodeToString(nonce), nil
}

// NewNote creates a new note and returns a unique human friendly ID.
// Data must be populated using UpdateNote.
func NewNote(title string) (string, error) {
	var note Note

	note.Title = title
	note.UploadTime = time.Now()

	for {
		var err error
		note.ID, err = idFromNote(note)
		if err != nil {
			return "", err
		}

		result, err := r.Table("notes").Insert(note).RunWrite(session)
		if err != nil {
			return "", err
		}

		if result.Errors == 0 {
			break
		}
	}

	return note.ID, nil
}

// UpdateNote updates the information of a note
func UpdateNote(note Note) error {
	note.UpdatedTime = time.Now()
	note.UploadTime = time.Time{}

	result, err := r.Table("notes").Get(note.ID).Update(note).RunWrite(session)
	if err != nil {
		return err
	}

	if result.Errors > 0 {
		return errors.New("airlift: " + result.FirstError)
	}

	return nil
}

// DeleteNote deletes a note
func DeleteNote(id string) error {
	result, err := r.Table("notes").Get(id).Delete().RunWrite(session)
	if err != nil {
		return err
	}

	if result.Errors > 0 {
		return errors.New("airlift: " + result.FirstError)
	}

	return nil
}

// SetNoteStar sets the status of a note's star
func SetNoteStar(id string, username string, starred bool) error {
	var query r.Term
	if starred {
		query = r.Table("notes").Get(id).Update(map[string]interface{}{
			"stars": r.Row.Field("stars").SetInsert(username),
		})
	} else {
		query = r.Table("notes").Get(id).Update(map[string]interface{}{
			"stars": r.Row.Field("stars").SetDifference([]string{username}),
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
