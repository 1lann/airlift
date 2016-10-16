package main

import (
	"errors"
	"net/http"

	"github.com/1lann/airlift/airlift"
	"github.com/1lann/airlift/fs"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

var errFileNotFound = errors.New("main: file not found")

type fileCard struct {
	Name string
	Size string
	URL  string
}

func init() {
	registers = append(registers, func(r *gin.RouterGroup, t multitemplate.Render) {
		r.GET("/download/:fileType/:id", func(c *gin.Context) {
			fileType := c.Param("fileType")
			id := c.Param("id")
			dl := c.Query("dl")

			filename, err := getUploadedFilename(fileType, id)
			if err == errFileNotFound {
				c.HTML(http.StatusNotFound, "not-found", nil)
				return
			}

			file, err := fs.RetrieveFile("airlift", fileType+"/"+id+".pdf")
			if err != nil {
				panic(err)
			}

			if dl == "force" {
				c.Header("Content-Type", "application/octet-stream")
				c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
			} else {
				c.Header("Content-Type", file.ContentType)
			}

			c.Header("ETag", file.ETag)
			http.ServeContent(c.Writer, c.Request, filename, file.LastModified, file)
		})
	})
}

func getUploadedFilename(fileType, id string) (string, error) {
	if fileType == "papers" || fileType == "solutions" || fileType == "sources" {
		return getPaperFilename(id, fileType)
	}

	if fileType == "notes" {
		return getNoteFilename(id)
	}

	return "", errFileNotFound
}

func getPaperFilename(id string, fileType string) (string, error) {
	paper, err := airlift.GetPaper(id)
	if err != nil {
		return "", err
	}

	if paper.Title == "" {
		return "", errFileNotFound
	}

	base := paper.Title + " - " + paper.Author
	if fileType == "solutions" {
		return base + " Solutions.pdf", nil
	} else if fileType == "sources" {
		return base + " Source Booklet.pdf", nil
	}

	return base + ".pdf", nil
}

func getNoteFilename(id string) (string, error) {
	note, err := airlift.GetNote(id)
	if err != nil {
		return "", err
	}

	if note.Title == "" {
		return "", errFileNotFound
	}

	return note.Title + " - " + note.Author + ".pdf", nil
}
