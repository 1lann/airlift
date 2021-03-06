package main

import (
	"net/http"
	"time"

	"github.com/1lann/airlift/airlift"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

type paperGroup struct {
	Year   int
	Papers []airlift.FullPaper
}

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
					Action:      "View subject",
					Link:        "/subjects/" + subject.ID,
				}
			}

			htmlOK(c, "subjects-list", gin.H{
				"ActiveMenu": "schedule",
				"Subjects":   subjectCards,
			})
		})

		t.AddFromFiles("subject", viewsPath+"/subject.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/subjects/:id", func(c *gin.Context) {
			id := c.Param("id")
			user := c.MustGet("user").(airlift.User)

			subject, err := airlift.GetFullSubject(id, user.Username)
			if err != nil {
				panic(err)
			}

			var starredNotes []airlift.Note
			var otherNotes []airlift.Note
			var uncompPapers []paperGroup
			var compPapers []paperGroup

			for _, note := range subject.Notes {
				if note.HasStarred {
					starredNotes = append(starredNotes, note)
				} else {
					otherNotes = append(otherNotes, note)
				}
			}

			for _, paper := range subject.Papers {
				if paper.HasCompleted {
					addAndGroup(&compPapers, airlift.FullPaper{Paper: paper})
				} else {
					addAndGroup(&uncompPapers, airlift.FullPaper{Paper: paper})
				}
			}

			if subject.Name == "" {
				c.HTML(http.StatusNotFound, "not-found", nil)
				return
			}

			htmlOK(c, "subject", gin.H{
				"ActiveMenu":  "schedule",
				"Subject":     subject,
				"Starred":     starredNotes,
				"OtherNotes":  otherNotes,
				"Completed":   compPapers,
				"OtherPapers": uncompPapers,
				"ExamTime":    formatScheduleTime(subject.ExamTime),
				"ExamPassed":  time.Now().After(subject.ExamTime),
			})
		})
	})
}

func addAndGroup(papers *[]paperGroup, paper airlift.FullPaper) {
	if len(*papers) == 0 ||
		(*papers)[len(*papers)-1].Year != paper.Year {
		*papers = append(*papers, paperGroup{
			Year:   paper.Year,
			Papers: []airlift.FullPaper{paper},
		})
	} else {
		(*papers)[len(*papers)-1].Papers = append(
			(*papers)[len(*papers)-1].Papers,
			paper,
		)
	}
}
