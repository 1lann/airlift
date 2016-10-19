package main

import (
	"net/http"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		r.POST("/delete/note", func(c *gin.Context) {
			id := c.PostForm("id")
			user := c.MustGet("user").(airlift.User)

			note, err := airlift.GetNote(id)
			if err != nil {
				panic(err)
			}

			if note.Uploader != user.Username {
				c.AbortWithStatus(http.StatusNotAcceptable)
				return
			}

			err = fs.DeleteFile("airlift", "notes/"+note.ID+".pdf")
			if err != nil {
				panic(err)
			}

			err = airlift.DeleteNote(id)
			if err != nil {
				panic(err)
			}

			session := sessions.Default(c)
			session.AddFlash("delete", "upload")
			session.Save()

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		r.POST("/delete/paper", func(c *gin.Context) {
			id := c.PostForm("id")
			user := c.MustGet("user").(airlift.User)

			paper, err := airlift.GetPaper(id)
			if err != nil {
				panic(err)
			}

			if paper.Uploader != user.Username {
				c.AbortWithStatus(http.StatusNotAcceptable)
				return
			}

			err = fs.DeleteFile("airlift", "papers/"+paper.ID+".pdf")
			if err != nil {
				panic(err)
			}

			if paper.SolutionsSize > 0 {
				err = fs.DeleteFile("airlift", "solutions/"+paper.ID+".pdf")
				if err != nil {
					panic(err)
				}
			}

			if paper.SourceSize > 0 {
				err = fs.DeleteFile("airlift", "sources/"+paper.ID+".pdf")
				if err != nil {
					panic(err)
				}
			}

			err = airlift.DeletePaper(id)
			if err != nil {
				panic(err)
			}

			session := sessions.Default(c)
			session.AddFlash("delete", "upload")
			session.Save()

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	})
}
