package util

import (
	"errors"
	"fmt"
	"log"

	"github.com/bogem/id3v2"
)

func UpdateTags(path string, title string, album string, artist string) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return errors.New(fmt.Sprintf("Error while opening mp3 file: %s", err))
	}
	defer tag.Close()

	tag.DeleteAllFrames()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)

	tag.SetTitle(title)
	tag.SetAlbum(album)
	tag.SetArtist(artist)

	if err = tag.Save(); err != nil {
		log.Printf("[Error] save tag error, %v", err)
		return errors.New(fmt.Sprintf("Error while saving a tag: %s", err))
	}

	return nil
}

func ReadTags(path string) (string, string, string, error) {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		log.Printf("[Error]读取Mp3文件失败, path = [%s], error = [%s]", path, err)
		return "", "", "", errors.New(fmt.Sprintf("Error while opening mp3 file: %s", err))
	}
	defer tag.Close()

	title := tag.Title()
	album := tag.Album()
	artist := tag.Artist()

	return title, album, artist, nil
}
