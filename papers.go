package main

import (
	"net/http"
	"sync"

	"github.com/1lann/airlift/airlift"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("papers", viewsPath+"/papers.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/papers", viewUserPapers)

		t.AddFromFiles("view-paper", viewsPath+"/view-paper.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/papers/:id", viewPaper)

		r.POST("/papers/:id/complete", func(c *gin.Context) {
			completed := c.PostForm("completed") == "true"
			username := c.MustGet("user").(airlift.User).Username

			err := airlift.SetPaperCompleted(c.Param("id"), username, completed)
			if err != nil {
				panic(err)
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	})
}

func viewPaper(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(airlift.User)

	paper, err := airlift.GetFullPaper(id, user.Username)
	if err != nil {
		panic(err)
	}

	if paper.Title == "" {
		c.HTML(http.StatusNotFound, "not-found", nil)
		return
	}

	files := []fileCard{
		{
			Name: "Question paper",
			Size: humanize.Bytes(paper.QuestionsSize),
			URL:  "/download/papers/" + paper.ID,
		},
	}

	if paper.SourceSize > 0 {
		files = append(files, fileCard{
			Name: "Source booklet",
			Size: humanize.Bytes(paper.SourceSize),
			URL:  "/download/sources/" + paper.ID,
		})
	}

	if paper.SolutionsSize > 0 {
		files = append(files, fileCard{
			Name: "Solutions",
			Size: humanize.Bytes(paper.SolutionsSize),
			URL:  "/download/solutions/" + paper.ID,
		})
	}

	session := sessions.Default(c)
	uploadFlashes := session.Flashes("upload")
	uploadSuccess := ""
	if len(uploadFlashes) > 0 {
		uploadSuccess = uploadFlashes[0].(string)
	}
	session.Save()

	htmlOK(c, "view-paper", gin.H{
		"ActiveMenu":    "papers",
		"Paper":         paper,
		"Files":         files,
		"IsUploader":    paper.Uploader == user.Username,
		"UploadTime":    formatBasicTime(paper.UploadTime),
		"UpdatedTime":   formatBasicTime(paper.UpdatedTime),
		"UploadSuccess": uploadSuccess,
	})
}

func viewUserPapers(c *gin.Context) {
	user := c.MustGet("user").(airlift.User)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var uploaded []paperGroup
	go func() {
		defer func() {
			wg.Done()
		}()
		var err error
		dbUploaded, err := airlift.GetUploadedPapers(user.Username)
		if err != nil {
			panic(err)
		}

		for _, upload := range dbUploaded {
			addAndGroup(&uploaded, upload)
		}
	}()

	var completed []paperGroup
	go func() {
		defer func() {
			wg.Done()
		}()
		var err error
		dbCompleted, err := airlift.GetCompletedPapers(user.Username)
		if err != nil {
			panic(err)
		}

		for _, complete := range dbCompleted {
			addAndGroup(&completed, complete)
		}
	}()

	deleted := false
	session := sessions.Default(c)
	uploadFlashes := session.Flashes("upload")
	if len(uploadFlashes) > 0 && uploadFlashes[0] == "delete" {
		deleted = true
	}
	session.Save()

	wg.Wait()

	htmlOK(c, "papers", gin.H{
		"ActiveMenu": "papers",
		"Completed":  completed,
		"Uploaded":   uploaded,
		"Deleted":    deleted,
	})
}
