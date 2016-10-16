package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"
	"github.com/gin-gonic/gin"
)

func parseNoteForm(c *gin.Context) (airlift.Note, error) {
	title := strings.TrimSpace(c.PostForm("title"))
	if !isTitleValid(title) {
		return airlift.Note{}, errors.New("upload: form validation error")

	}

	author := strings.TrimSpace(c.PostForm("author"))
	if author == "" {
		return airlift.Note{}, errors.New("upload: form validation error")

	}

	subject := c.PostForm("subject")
	if !isSubject(subject, c.MustGet("user").(airlift.User)) {
		return airlift.Note{}, errors.New("upload: form validation error")
	}

	return airlift.Note{
		Title:    title,
		Public:   true,
		Author:   author,
		Subject:  subject,
		Uploader: c.MustGet("user").(airlift.User).Username,
	}, nil
}

func uploadNote(c *gin.Context) {
	note, err := parseNoteForm(c)
	if err != nil {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}

	// hasFile := false
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}

		// hasFile = true
	}

	// TODO: Check for `update`, validate permissions and compare to hasFile.
	// Otherwise NotAcceptable

	id, err := airlift.NewNote(note)
	if err != nil {
		panic(err)
	}

	err = fs.UploadFile("airlift", "notes/"+id+".pdf", "application/pdf", file)
	if err == fs.ErrTooBig {
		c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		return
	} else if err == fs.ErrInvalidType {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
