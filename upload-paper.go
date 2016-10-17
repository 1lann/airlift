package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/1lann/airlift/airlift"
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

	update := c.PostForm("update")

	if update != "" {
		var dbPaper airlift.Paper
		dbPaper, err = airlift.GetPaper(update)
		if err != nil {
			panic(err)
		}

		if dbPaper.Author != paper.Author {
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}

		paper.ID = dbPaper.ID
	} else {
		var id string
		id, err = airlift.NewPaper(paper.Title)
		if err != nil {
			panic(err)
		}

		paper.ID = id
	}

	wg := new(sync.WaitGroup)
	uploadPaperFiles(paper.ID, wg, c, &paper)

	if update == "" && paper.QuestionsSize == 0 {
		airlift.DeletePaper(paper.ID)
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}

	wg.Wait()
	if c.IsAborted() {
		airlift.DeletePaper(paper.ID)
		return
	}

	err = airlift.UpdatePaper(paper)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"id": paper.ID,
	})
}

func uploadPaperFiles(id string, wg *sync.WaitGroup, c *gin.Context, paper *airlift.Paper) {
	source, _, err := c.Request.FormFile("source")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			paper.SourceSize = uploadFile(id, source, "sources", c)
		}()
	}

	solutions, _, err := c.Request.FormFile("solutions")
	if err != nil {
		if err != http.ErrMissingFile {
			panic(err)
		}
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			paper.SolutionsSize = uploadFile(id, solutions, "solutions", c)
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
			paper.QuestionsSize = uploadFile(id, paperFile, "papers", c)
		}()
	}
}
