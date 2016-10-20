package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/1lann/airlift/airlift"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func formatBasicTime(t time.Time) string {
	return getDay(t) + " " + t.Format("January 2006 at 3:04 PM")
}

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("notes", viewsPath+"/notes.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/notes", viewUserNotes)

		t.AddFromFiles("view-note", viewsPath+"/view-note.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/notes/:id", viewNote)

		r.POST("/notes/:id/star", func(c *gin.Context) {
			starred := c.PostForm("starred") == "true"
			username := c.MustGet("user").(airlift.User).Username

			err := airlift.SetNoteStar(c.Param("id"), username, starred)
			if err != nil {
				panic(err)
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	})
}

func viewNote(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(airlift.User)

	note, err := airlift.GetFullNote(id, user.Username)
	if err != nil {
		panic(err)
	}

	if note.Title == "" {
		c.HTML(http.StatusNotFound, "not-found", nil)
		return
	}

	files := []fileCard{
		{
			Name: "Notes",
			Size: humanize.Bytes(note.Size),
			URL:  "/download/notes/" + note.ID,
		},
	}

	session := sessions.Default(c)
	uploadFlashes := session.Flashes("upload")
	uploadSuccess := ""
	if len(uploadFlashes) > 0 {
		uploadSuccess = uploadFlashes[0].(string)
	}
	session.Save()

	htmlOK(c, "view-note", gin.H{
		"ActiveMenu":    "notes",
		"Note":          note,
		"Files":         files,
		"IsUploader":    note.Uploader == user.Username,
		"UploadTime":    formatBasicTime(note.UploadTime),
		"UpdatedTime":   formatBasicTime(note.UpdatedTime),
		"UploadSuccess": uploadSuccess,
	})
}

func viewUserNotes(c *gin.Context) {
	user := c.MustGet("user").(airlift.User)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var starred []airlift.Note
	go func() {
		defer func() {
			wg.Done()
		}()
		var err error
		starred, err = airlift.GetStarredNotes(user.Username)
		if err != nil {
			panic(err)
		}
	}()

	var uploaded []airlift.Note
	go func() {
		defer func() {
			wg.Done()
		}()
		var err error
		uploaded, err = airlift.GetUploadedNotes(user.Username)
		if err != nil {
			panic(err)
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

	htmlOK(c, "notes", gin.H{
		"ActiveMenu": "notes",
		"Starred":    starred,
		"Uploaded":   uploaded,
		"Deleted":    deleted,
	})
}
