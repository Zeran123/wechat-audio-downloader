/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ZeRanW/wechat-audio-downloader/internal/util"
	"github.com/anaskhan96/soup"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动Web服务",
	Run: func(cmd *cobra.Command, args []string) {
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

			files := make([]string, 0)

			doc := soup.HTMLParse(content)
			audioHtml := doc.FindAll("mpvoice", "class", "audio_iframe")
			for _, tag := range audioHtml {
				fileId := tag.Attrs()["voice_encode_fileid"]
				fileName := tag.Attrs()["name"]

				// https: //res.wx.qq.com/voice/getvoice?mediaid=
				aResp, err := http.Get(fmt.Sprintf("https://res.wx.qq.com/voice/getvoice?mediaid=%s", fileId))
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
				fullPath := fmt.Sprintf("%s/%s.mp3", dir, fileName)
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

				util.UpdateTags(fullPath, fileName, album, artist)
				t, a, ar, err := util.ReadTags(fullPath)
				if err != nil {
					log.Printf("[Error] 读取Mp3标签失败, error = [%s]", err)
					c.JSON(500, gin.H{
						"message": err,
					})
				}

				log.Printf("[Info] 已更新文件Mp3标签, path = [%s], title = [%s], album = [%s], artist = [%s]", fullPath, t, a, ar)
				log.Printf("[Info] 音频文件保存成功, name = [%s], path = [%s]", fileName, fullPath)

				files = append(files, fileName)
			}

			c.JSON(200, gin.H{
				"message": strings.Join(files, ", "),
			})
		})
		r.Run()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
