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
	"os"
	"path/filepath"
	"strings"

	"github.com/ZeRanW/wechat-audio-downloader/internal/util"
	"github.com/spf13/cobra"
	"github.com/wellmoon/m4aTag/mtag"
)

var title string
var album string
var artist string

// mp3Cmd represents the mp3 command
var mp3Cmd = &cobra.Command{
	Use:   "mp3 [OPTIONS] [FILE]",
	Short: "更新和读取Mp3文件的标签",
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		finfo, err := os.Stat(path)
		if err != nil {
			fmt.Printf("读取文件失败, path = [%s], error = [%v]\n", path, err)
			return
		}

		if !finfo.IsDir() {
			updateFile(finfo.Name(), path)
		} else {
			// 目录
			dirs, err := os.ReadDir(path)
			if err != nil {
				fmt.Printf("[Error] 读取目录内容失败, error = [%s]\n", err)
				return
			}
			for _, dir := range dirs {
				info, _ := dir.Info()
				updateFile(info.Name(), fmt.Sprintf("%s%s", path, info.Name()))
			}
		}

	},
}

func updateFile(fileName string, path string) {
	fmt.Printf("[Info] 开始更新文件 -> %s\n", path)
	// 单个文件
	titleTag := ""
	ext := filepath.Ext(fileName)
	if title == "{FILE_NAME}" {
		if ext != ".mp3" && ext != ".m4a" {
			return
		}
		titleTag = strings.Replace(fileName, ext, "", -1)
	} else {
		titleTag = title
	}
	if ext == ".mp3" {
		util.UpdateTags(path, titleTag, album, artist)
		t, a, aa, err := util.ReadTags(path)
		if err != nil {
			fmt.Printf("[Error] 读取Mp3标签失败, error = [%v]\n", err)
			return
		}
		fmt.Printf("[Info] 更新后的MP3标签 -> title = [%s], ablum = [%s], artist = [%s]\n", t, a, aa)
	} else if ext == ".m4a" {

		mtag.UpdateM4aTag(true, path, titleTag, artist, album, "", "")
		tag, err := mtag.ReadM4aTag(path)
		if err != nil {
			fmt.Printf("[Error] 读取M4a标签失败, error = [%v]\n", err)
			return
		}
		fmt.Printf("[Info] 更新后的M4a标签 -> title = [%s], ablum = [%s], artist = [%s]\n", tag.Name, tag.Album, tag.Artist)
	}
}

func init() {
	rootCmd.AddCommand(mp3Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mp3Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mp3Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	mp3Cmd.Flags().StringVar(&album, "album", "我的故事", "设置专辑")
	mp3Cmd.Flags().StringVar(&artist, "artist", "来自网络", "设置参与艺术家")
	mp3Cmd.Flags().StringVar(&title, "title", "{FILE_NAME}", "设置标题")
}
