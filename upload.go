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
		r.GET("/upload/note", func(c *gin.Context) {
			htmlOK(c, "upload-note", gin.H{
				"ActiveMenu": "notes",
				"Update":     c.Query("update"),
			})
		})

		t.AddFromFiles("upload-paper", viewsPath+"/upload-paper.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/upload/paper", func(c *gin.Context) {
			// TODO: Pre fill both forms
			htmlOK(c, "upload-paper", gin.H{
				"ActiveMenu": "papers",
				"Update":     c.Query("update"),
			})
		})

		r.POST("/upload/note", uploadNote)
		r.POST("/upload/paper", uploadPaper)
	})
}
