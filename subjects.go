package main

import (
	"net/http"

	"github.com/1lann/airlift/airlift"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("subjects-list", viewsPath+"/subjects-list.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/subjects", func(c *gin.Context) {
			subjects, err := airlift.AllSubjects()
			if err != nil {
				panic(err)
			}

			subjectCards := make([]listCard, len(subjects))

			for i, subject := range subjects {
				subjectCards[i] = listCard{
					Header:      subject.Name,
					Description: formatScheduleTime(subject.ExamTime),
					Action:      "View info",
					Link:        "/subjects/" + subject.ID,
				}
			}

			htmlOK(c, "subjects-list", gin.H{
				"ActiveMenu": "subjects",
				"Subjects":   subjectCards,
			})
		})

		t.AddFromFiles("subject", viewsPath+"/subject.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/subjects/:id", func(c *gin.Context) {
			id := c.Param("id")

			subject, err := airlift.GetFullSubject(id)
			if err != nil {
				panic(err)
			}

			if subject.Name == "" {
				c.HTML(http.StatusNotFound, "not-found", nil)
				return
			}

			htmlOK(c, "subject", gin.H{
				"ActiveMenu": "subjects",
			})
		})
	})
}
