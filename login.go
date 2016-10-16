package main

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/app4"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

const hashCost = 13

type showMessage struct {
	Title   string
	Message string
	Type    string
}

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("login", viewsPath+"/login.tmpl",
			viewsPath+"/components/base.tmpl")

		r.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			flashes := session.Flashes("login")
			session.Save()
			if len(flashes) > 0 {
				c.HTML(http.StatusOK, "login", gin.H{"Error": flashes[0]})
				return
			}
			c.HTML(http.StatusOK, "login", gin.H{})
		})

		r.POST("/", handleLogin)

		r.GET("/logout", func(c *gin.Context) {
			session := sessions.Default(c)
			session.AddFlash(showMessage{
				Title:   "You have been logged out",
				Message: "Thanks for using Airlift.",
				Type:    "info",
			}, "login")
			session.Delete("username")
			session.Save()
			c.Redirect(http.StatusSeeOther, "/")
		})
	})
}

func handleLogin(c *gin.Context) {
	wait := time.After(time.Second * 3)
	username := c.PostForm("username")
	password := c.PostForm("password")
	user, err := airlift.GetUser(username)
	if err == airlift.ErrNotFound {
		<-wait
		c.HTML(http.StatusOK, "login", gin.H{
			"Error": showMessage{
				Title: "Student ID does not exist",
				Message: "Make sure you entered your student ID correctly. " +
					"You must also be in year 12 to be able to use Airlift.",
				Type: "error",
			},
		})
		return
	} else if err != nil {
		log.Println("login:", err)
		c.HTML(http.StatusOK, "login", gin.H{
			"Error": showMessage{
				Title:   "Database error",
				Message: "Sorry, there was an error querying the database. Contact Chuie for help.",
				Type:    "error",
			},
		})
		return
	}

	if user.Password != "" {
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err == nil {
			<-wait
			session := sessions.Default(c)
			session.Set("username", username)
			session.Save()
			c.Redirect(http.StatusSeeOther, "/schedule")
			return
		}
	}

	app4s, err := app4.Login(username, password)
	if err == app4.ErrInvalidCredentials {
		<-wait
		c.HTML(http.StatusOK, "login", gin.H{
			"Error": showMessage{
				Title: "Incorrect password",
				Message: "Airlift uses the same password " +
					"you use to login to school Wi-Fi, the Intranet, and your school email.",
				Type: "error",
			},
		})
		return
	} else if err != nil {
		log.Println("login:", err)
		c.HTML(http.StatusOK, "login", gin.H{
			"Error": showMessage{
				Title: "School server error",
				Message: "Sorry, there was an error contacting the school's " +
					"servers. Please try again later.",
				Type: "error",
			},
		})
		return
	}

	err = registerUser(user, password, app4s)
	if err != nil {
		log.Println("login:", err)
		<-wait
		c.HTML(http.StatusOK, "login", gin.H{
			"Error": showMessage{
				Title:   "Server error",
				Message: "Sorry, there was an error while logging you in for the first time. Contact me@chuie.io for help.",
				Type:    "error",
			},
		})
		return
	}

	<-wait
	session := sessions.Default(c)
	session.Set("username", username)
	session.Save()
	c.Redirect(http.StatusSeeOther, "/schedule")
}

func registerUser(user airlift.User, password string, app4s app4.Session) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return err
	}

	if len(user.Schedule) == 0 {
		// Also get schedule
		var classes []string
		classes, err = app4s.GetClassNames()
		if err != nil {
			return err
		}

		err = airlift.UpdateScheduleByMatch(user.Username, classes)
		if err != nil {
			return err
		}
	}

	err = airlift.UpdatePassword(user.Username, string(hash))
	if err != nil {
		log.Println("login:", err)
		return err
	}

	return nil
}
