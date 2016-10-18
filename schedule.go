package main

import (
	"strconv"
	"time"

	"github.com/1lann/airlift/airlift"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

func getDay(t time.Time) string {
	num := strconv.Itoa(t.Day())
	switch t.Day() {
	case 1, 21, 31:
		return num + "st"
	case 2, 22, 32:
		return num + "nd"
	case 3, 23, 33:
		return num + "rd"
	default:
		return num + "th"
	}
}

func formatScheduleTime(t time.Time) string {
	return t.Weekday().String() + ", " + getDay(t) + " " + t.Month().String() +
		" at " + t.Format("3:04 PM")
}

type listCard struct {
	Header      string
	Description string
	Action      string
	Link        string
	Icon        string
	Color       string
}

func scheduleHandler(c *gin.Context) {
	user := c.MustGet("user").(airlift.User)
	schedule, err := airlift.GetScheduleFor(user)
	if err != nil {
		panic(err)
	}

	var pastSubjects []listCard
	var upcomingSubjects []listCard

	for _, subject := range schedule {
		if time.Now().After(subject.ExamTime) {
			pastSubjects = append(pastSubjects, listCard{
				Header:      subject.Name,
				Description: formatScheduleTime(subject.ExamTime),
				Action:      "View info",
				Link:        "/subjects/" + subject.ID,
			})
		} else if time.Now().YearDay() == subject.ExamTime.YearDay() {
			upcomingSubjects = append(upcomingSubjects, listCard{
				Header: "Good luck",
				Description: "Your " + subject.Name +
					" exam starts at " + subject.ExamTime.Format("3:04 PM"),
				Action: "Cram now",
				Icon:   "smile",
				Color:  "green",
				Link:   "/subjects/" + subject.ID,
			})
		} else if time.Now().YearDay()+1 == subject.ExamTime.YearDay() {
			upcomingSubjects = append(upcomingSubjects, listCard{
				Header:      subject.Name + " is your next exam",
				Description: formatScheduleTime(subject.ExamTime),
				Action:      "Prepare now",
				Icon:        "info",
				Color:       "teal",
				Link:        "/subjects/" + subject.ID,
			})
		} else {
			upcomingSubjects = append(upcomingSubjects, listCard{
				Header:      subject.Name,
				Description: formatScheduleTime(subject.ExamTime),
				Action:      "Prepare now",
				Link:        "/subjects/" + subject.ID,
			})
		}
	}

	// Reverses pastSubjects
	for left, right := 0, len(pastSubjects)-1; left < right; left, right = left+1, right-1 {
		pastSubjects[left], pastSubjects[right] = pastSubjects[right], pastSubjects[left]
	}

	htmlOK(c, "schedule", gin.H{
		"ActiveMenu": "schedule",
		"Past":       pastSubjects,
		"Upcoming":   upcomingSubjects,
	})
}

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("schedule", viewsPath+"/schedule.tmpl",
			viewsPath+"/components/base.tmpl")
		r.GET("/schedule", scheduleHandler)
	})
}
