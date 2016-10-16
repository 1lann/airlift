package main

import (
	"net/http"

	"github.com/1lann/airlift/airlift"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("papers", viewsPath+"/papers.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/papers", func(c *gin.Context) {
			htmlOK(c, "papers", gin.H{
				"ActiveMenu": "papers",
			})
		})

		t.AddFromFiles("view-paper", viewsPath+"/view-paper.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/papers/:id", func(c *gin.Context) {
			id := c.Param("id")
			paper, err := airlift.GetPaper(id)
			if err != nil {
				panic(err)
			}

			if paper.Title == "" {
				c.HTML(http.StatusNotFound, "not-found", nil)
				return
			}

			htmlOK(c, "view-paper", gin.H{
				"ActiveMenu": "papers",
			})
		})
	})
}
