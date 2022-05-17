package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/anaskhan96/soup"
	"github.com/bogem/id3v2"
	"github.com/gin-gonic/gin"
)

func main() {
	path := os.Getenv("DOWNLOAD_PATH")
	u := os.Getenv("USER")
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

		filePath := fmt.Sprintf("%s/%s.mp3", path, name)
		file, err := os.Create(filePath)
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

		updateMp3Tag(filePath, name)

		group, err := user.Lookup(u)
		uid, _ := strconv.Atoi(group.Uid)
		gid, _ := strconv.Atoi(group.Gid)

		err = syscall.Chown(filePath, uid, gid)

		c.JSON(200, gin.H{
			"message": name,
		})
	})
	r.Run()
}

func updateMp3Tag(path string, title string) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return errors.New(fmt.Sprintf("Error while opening mp3 file: %s", err))
	}
	defer tag.Close()

	tag.SetTitle(title)
	tag.SetAlbum("联合天畅")
	tag.SetArtist("联合天畅")

	comment := id3v2.CommentFrame{
		Encoding: id3v2.EncodingUTF8,
	}
	tag.AddCommentFrame(comment)

	if err = tag.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error while saving a tag: %s", err))
	}

	return nil
}
