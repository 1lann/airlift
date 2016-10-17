package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/1lann/airlift/airlift"
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

	processNoteUpdate(c, &note)
	if c.IsAborted() {
		return
	}

	err = airlift.UpdateNote(note)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"id": note.ID,
	})
}

func processNoteUpdate(c *gin.Context, note *airlift.Note) {
	hasFile := true
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}

		hasFile = false
	}

	update := c.PostForm("update")

	if update != "" {
		var dbNote airlift.Note
		dbNote, err := airlift.GetNote(update)
		if err != nil {
			panic(err)
		}

		if dbNote.Author != note.Author {
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}

		note.ID = dbNote.ID
	} else if !hasFile {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	} else {
		var err error
		note.ID, err = airlift.NewNote(note.Title)
		if err != nil {
			panic(err)
		}
	}

	if hasFile {
		note.Size = uploadFile(note.ID, file, "notes", c)
		if c.IsAborted() {
			airlift.DeleteNote(note.ID)
			return
		}
	}
}