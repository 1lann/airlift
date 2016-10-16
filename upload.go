package main

import (
	"strings"

	"github.com/1lann/airlift/airlift"
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

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("upload-note", viewsPath+"/upload-note.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/upload/note", func(c *gin.Context) {
			htmlOK(c, "upload-note", gin.H{
				"ActiveMenu": "notes",
			})
		})

		t.AddFromFiles("upload-paper", viewsPath+"/upload-paper.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/upload/paper", func(c *gin.Context) {
			htmlOK(c, "upload-paper", gin.H{
				"ActiveMenu": "papers",
			})
		})

		r.POST("/upload/note", uploadNote)
		r.POST("/upload/paper", uploadPaper)
	})
}
