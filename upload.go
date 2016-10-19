package main

import (
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

func isTitleValid(name string) bool {
	if len(name) < 3 {
		return false
	}

	if strings.ContainsAny(name, ":/\\<>\"|?*") {
		return false
	}

	return true
}

func titleCase(title string) string {
	words := strings.Split(title, " ")
	construct := ""
	for _, word := range words {
		if len(word) > 3 {
			construct += strings.ToTitle(word[:1]) + word[1:] + " "
		}
	}

	return construct[:len(construct)-1]
}

func isSubject(subject string, user airlift.User) bool {
	for _, class := range user.Schedule {
		if class == subject {
			return true
		}
	}

	return false
}

func uploadFile(id string, file multipart.File, fileType string, c *gin.Context) uint64 {
	n, err := fs.UploadFile("airlift", fileType+"/"+id+".pdf", "application/pdf", file)
	if err == fs.ErrTooBig {
		c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		return 0
	} else if err == fs.ErrInvalidType {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return 0
	} else if err != nil {
		panic(err)
	}

	return n
}

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("upload-note", viewsPath+"/upload-note.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/upload/note", viewUploadNote)

		t.AddFromFiles("upload-paper", viewsPath+"/upload-paper.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/upload/paper", viewUploadPaper)

		r.POST("/upload/note", uploadNote)
		r.POST("/upload/paper", uploadPaper)
	})
}

func viewUploadNote(c *gin.Context) {
	id := c.Query("update")

	var note airlift.Note
	if id != "" {
		var err error
		note, err = airlift.GetNote(id)
		if err != nil {
			panic(err)
		}

		if note.Title == "" {
			c.Redirect(http.StatusSeeOther, "/upload/note")
			return
		}
	}

	user := c.MustGet("user").(airlift.User)

	filledSubject := c.Query("subject")
	subjectFound := false
	for _, subject := range user.Schedule {
		if filledSubject == subject {
			subjectFound = true
			break
		}
	}

	if !subjectFound {
		filledSubject = ""
	}

	subjects, err := airlift.GetAlphaScheduleFor(user)
	if err != nil {
		panic(err)
	}

	htmlOK(c, "upload-note", gin.H{
		"ActiveMenu":    "notes",
		"Update":        id,
		"Note":          note,
		"Uploader":      user.Name,
		"Subjects":      subjects,
		"FilledSubject": filledSubject,
	})
}

func viewUploadPaper(c *gin.Context) {
	id := c.Query("update")

	var paper airlift.Paper
	if id != "" {
		var err error
		paper, err = airlift.GetPaper(id)
		if err != nil {
			panic(err)
		}

		if paper.Title == "" {
			c.Redirect(http.StatusSeeOther, "/upload/paper")
			return
		}
	}

	user := c.MustGet("user").(airlift.User)

	filledSubject := c.Query("subject")
	subjectFound := false
	for _, subject := range user.Schedule {
		if filledSubject == subject {
			subjectFound = true
			break
		}
	}

	if !subjectFound {
		filledSubject = ""
	}

	subjects, err := airlift.GetAlphaScheduleFor(user)
	if err != nil {
		panic(err)
	}

	htmlOK(c, "upload-paper", gin.H{
		"ActiveMenu":    "papers",
		"Update":        id,
		"Paper":         paper,
		"Uploader":      user.Name,
		"Subjects":      subjects,
		"FilledSubject": filledSubject,
	})
}
