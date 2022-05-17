package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
	"github.com/gin-gonic/gin"
)

func main() {
	path := os.Getenv("DOWNLOAD_PATH")
	if path == "" {
		path = "."
	}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "up",
		})
	})
	r.POST("/download", func(c *gin.Context) {
		url := c.PostForm("url")
		// name := c.PostForm("name")

		resp, err := http.Get(url)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}

		content := string(body)
		defer resp.Body.Close()

		doc := soup.HTMLParse(content)
		name := doc.Find("h1", "id", "activity-name").Text()
		name = strings.TrimSpace(name)
		audio_html := doc.FindAll("mpvoice", "class", "audio_iframe")
		file_id := audio_html[0].Attrs()["voice_encode_fileid"]
		// https: //res.wx.qq.com/voice/getvoice?mediaid=
		aResp, err := http.Get(fmt.Sprintf("https://res.wx.qq.com/voice/getvoice?mediaid=%s", file_id))
		if err != nil {
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}
		defer aResp.Body.Close()

		file, err := os.Create(fmt.Sprintf("%s/%s.mp3", path, name))
		if err != nil {
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}
		_, err = io.Copy(file, aResp.Body)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}

		c.JSON(200, gin.H{
			"message": name,
		})
	})
	r.Run()
}
