package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

func formatBasicTime(t time.Time) string {
	return getDay(t) + " " + t.Month().String() + " " + strconv.Itoa(t.Year())
}

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("notes", viewsPath+"/notes.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/notes", func(c *gin.Context) {
			htmlOK(c, "notes", gin.H{
				"ActiveMenu": "notes",
			})
		})

		t.AddFromFiles("view-note", viewsPath+"/view-note.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/notes/:id", viewNote)
	})
}

func viewNote(c *gin.Context) {
	id := c.Param("id")
	note, err := airlift.GetFullNote(id)
	if err != nil {
		panic(err)
	}

	if note.Title == "" {
		c.HTML(http.StatusNotFound, "not-found", nil)
		return
	}

	stat, err := fs.StatFile("airlift", "notes/"+note.ID+".pdf")
	if err != nil {
		panic(err)
	}

	user := c.MustGet("user").(airlift.User)

	hasStarred := false
	for _, star := range note.Stars {
		if star == user.Username {
			hasStarred = true
		}
	}

	files := []fileCard{
		{
			Name: "Notes",
			Size: humanize.Bytes(uint64(stat.ObjectInfo.Size)),
			URL:  "/download/notes/" + note.ID,
		},
	}

	htmlOK(c, "view-note", gin.H{
		"ActiveMenu":  "notes",
		"Note":        note,
		"HasStarred":  hasStarred,
		"Files":       files,
		"IsAuthor":    note.Author == user.Username,
		"UploadDate":  formatBasicTime(note.UploadTime),
		"UpdatedDate": formatBasicTime(note.UpdatedTime),
	})
}
