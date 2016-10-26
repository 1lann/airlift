package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"

	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

var packagePath = os.Getenv("GOPATH") + "/src/github.com/1lann/airlift"
var viewsPath = packagePath + "/views"

var registers []func(r *gin.RouterGroup, t multitemplate.Render)

func main() {
	err := airlift.Connect(dbConnectOpts)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	err = fs.Connect(minioAddr, minioAccessKey, minioSecretKey)
	if err != nil {
		log.Fatal("failed to connect to object store: ", err)
	}

	r := gin.Default()
	t := multitemplate.New()

	registerBaseHandlers(r, t)

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options(sessionOpts)
	r.Use(sessions.Sessions("airlift", store))

	rg := r.Group("/")
	rg.Use(authMiddleware)

	t.AddFromFiles("not-found", viewsPath+"/not-found.tmpl",
		viewsPath+"/components/base.tmpl")

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "not-found", nil)
	})

	for _, registerFunc := range registers {
		registerFunc(rg, t)
	}
	r.HTMLRender = t
	r.Run()
}

func authMiddleware(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/static") ||
		c.Request.URL.Path == "/favicon.ico" {
		return
	}

	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok && c.Request.URL.Path == "/" {
		c.Next()
		return
	}

	if !ok {
		session.AddFlash(showMessage{
			Title:   "You aren't logged in",
			Message: "You need to log in first before you can see this content.",
			Type:    "error",
		}, "login")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/")
		c.Abort()
		return
	}

	if c.Request.URL.Path == "/" {
		c.Redirect(http.StatusSeeOther, "/schedule")
		c.Abort()
		return
	}

	user, err := airlift.GetUser(username)
	if err == airlift.ErrNotFound {
		session.AddFlash(showMessage{
			Title:   "Your account has gone missing",
			Message: "Was it deleted? Contact Chuie for help.",
			Type:    "error",
		}, "login")
		session.Delete("username")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/")
		c.Abort()
		return
	} else if err != nil {
		panic(err)
	}

	c.Set("user", user)
}

func registerBaseHandlers(r *gin.Engine, t multitemplate.Render) {
	r.Use(secure.Secure(secureOpts))

	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Recover from panic
				stackTrace := strings.Replace(string(debug.Stack()),
					os.Getenv("GOPATH")+"/src/", "", -1)
				log.Println("Recovered from panic:", fmt.Sprintf("%v\n", err)+stackTrace)
				c.HTML(http.StatusInternalServerError, "error", gin.H{
					"StackTrace": fmt.Sprintf("%v\n", err) + stackTrace,
				})
			}
		}()

		c.Next()
	})
}

func getGreeting() string {
	hour := time.Now().Hour()
	if hour < 5 {
		return "evening"
	} else if hour < 12 {
		return "morning"
	} else if hour < 18 {
		return "afternoon"
	}

	return "evening"
}

func htmlOK(c *gin.Context, template string, h map[string]interface{}) {
	h["User"] = c.MustGet("user")
	h["Greeting"] = getGreeting()
	c.HTML(http.StatusOK, template, h)
}

func init() {
	gob.Register(showMessage{})

	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		rg := r.Group("/static")
		rg.Use(func(c *gin.Context) {
			c.Header("Cache-Control", "max-age=10000000")
		})
		rg.Static("/", packagePath+"/static")
	})

	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		t.AddFromFiles("error", viewsPath+"/error.tmpl",
			viewsPath+"/components/base.tmpl")
	})
}
