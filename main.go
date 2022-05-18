package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
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
		name := c.PostForm("name")
		album := c.PostForm("album")
		artist := c.PostForm("artist")

		if album == "" {
			album = "我的故事"
		}

		if artist == "" {
			artist = "来自网络"
		}

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[Error] 打开公众号页面失败, url = [%s], error = [%s]", url, err)
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[Error] 获取页面源码失败, error = [%s]", err)
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}

		content := string(body)
		defer resp.Body.Close()

		doc := soup.HTMLParse(content)
		if name == "" {
			name = doc.Find("h1", "id", "activity-name").Text()
		}
		name = strings.TrimSpace(name)
		audio_html := doc.FindAll("mpvoice", "class", "audio_iframe")
		file_id := audio_html[0].Attrs()["voice_encode_fileid"]
		// https: //res.wx.qq.com/voice/getvoice?mediaid=
		aResp, err := http.Get(fmt.Sprintf("https://res.wx.qq.com/voice/getvoice?mediaid=%s", file_id))
		if err != nil {
			log.Printf("[Error] 获取音频内容失败, error = [%s]", err)
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}
		defer aResp.Body.Close()

		dir := fmt.Sprintf("%s/%s/%s", path, album, artist)
		os.MkdirAll(dir, os.FileMode(0775))
		fullPath := fmt.Sprintf("%s/%s.mp3", dir, name)
		file, err := os.Create(fullPath)
		if err != nil {
			log.Printf("[Error] 创建本地文件失败, path = [%s], error = [%s]", fullPath, err)
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}
		_, err = io.Copy(file, aResp.Body)
		if err != nil {
			log.Printf("[Error] 写入本地文件失败, error = [%s]", err)
			c.JSON(500, gin.H{
				"message": err,
			})
			return
		}

		updateMp3Tag(fullPath, name, album, artist)

		if runtime.GOOS == "linux" {
			log.Printf("[Info] 开始更新文件拥有者为 user = [%s], path = [%s]", u, fullPath)
			group, _ := user.Lookup(u)
			uid, _ := strconv.Atoi(group.Uid)
			gid, _ := strconv.Atoi(group.Gid)
			_ = syscall.Chown(fullPath, uid, gid)
		}

		log.Printf("[Info] 音频文件保存成功, name = [%s], path = [%s]", name, fullPath)
		c.JSON(200, gin.H{
			"message": name,
		})
	})
	r.Run()
}

func updateMp3Tag(path string, title string, album string, artist string) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return errors.New(fmt.Sprintf("Error while opening mp3 file: %s", err))
	}
	defer tag.Close()

	tag.SetTitle(title)
	tag.SetAlbum(album)
	tag.SetArtist(artist)

	comment := id3v2.CommentFrame{
		Encoding: id3v2.EncodingUTF8,
	}
	tag.AddCommentFrame(comment)

	if err = tag.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error while saving a tag: %s", err))
	}

	return nil
}
