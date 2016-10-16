package main

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"
	"github.com/gin-gonic/gin"
)

func parsePaperForm(c *gin.Context) (airlift.Paper, error) {
	title := strings.TrimSpace(c.PostForm("title"))
	if !isTitleValid(title) {
		return airlift.Paper{}, errors.New("upload: form validation error")
	}

	author := strings.TrimSpace(c.PostForm("author"))
	if author == "" {
		return airlift.Paper{}, errors.New("upload: form validation error")
	}

	year, err := strconv.Atoi(strings.TrimSpace(c.PostForm("year")))
	if err != nil {
		return airlift.Paper{}, errors.New("upload: form validation error")
	}

	if year < 1990 || year > 2016 {
		return airlift.Paper{}, errors.New("upload: form validation error")
	}

	subject := c.PostForm("subject")
	if !isSubject(subject, c.MustGet("user").(airlift.User)) {
		return airlift.Paper{}, errors.New("upload: form validation error")
	}

	return airlift.Paper{
		Title:    title,
		Year:     year,
		Public:   true,
		Author:   author,
		Subject:  subject,
		Uploader: c.MustGet("user").(airlift.User).Username,
	}, nil
}

func uploadPaper(c *gin.Context) {
	paper, err := parsePaperForm(c)
	if err != nil {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}

	var id string

	if c.PostForm("update") == "" {
		id, err = airlift.NewPaper(paper)
		if err != nil {
			panic(err)
		}
	} else {
		// TODO: Check paper exists and permissions, if not then NotAcceptable
	}

	wg := new(sync.WaitGroup)
	uploadPaperFiles(id, wg, c, &paper)

	// TODO: Update database

	wg.Wait()

	if c.IsAborted() {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func uploadPaperFiles(id string, wg *sync.WaitGroup, c *gin.Context, paper *airlift.Paper) {
	source, _, err := c.Request.FormFile("source")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}
	} else {
		paper.HasSource = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			uploadSource(id, source, c)
		}()
	}

	solutions, _, err := c.Request.FormFile("solutions")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}
	} else {
		paper.HasSolutions = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			uploadSolutions(id, solutions, c)
		}()
	}

	paperFile, _, err := c.Request.FormFile("questions")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			uploadQuestions(id, paperFile, c)
		}()
	}
}

func uploadQuestions(id string, paper multipart.File, c *gin.Context) {
	err := fs.UploadFile("airlift", "papers/"+id+".pdf", "application/pdf", paper)
	if err == fs.ErrTooBig {
		c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		return
	} else if err == fs.ErrInvalidType {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		panic(err)
	}
}

func uploadSolutions(id string, solutions multipart.File, c *gin.Context) {
	err := fs.UploadFile("airlift", "solutions/"+id+".pdf", "application/pdf", solutions)
	if err == fs.ErrTooBig {
		c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		return
	} else if err == fs.ErrInvalidType {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		panic(err)
	}
}

func uploadSource(id string, solutions multipart.File, c *gin.Context) {
	err := fs.UploadFile("airlift", "sources/"+id+".pdf", "application/pdf", solutions)
	if err == fs.ErrTooBig {
		c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		return
	} else if err == fs.ErrInvalidType {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		panic(err)
	}
}
