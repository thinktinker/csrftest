package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Image is not stored in the database
type Image struct {
	GalleryID uint
	Filename  string
}

// func (i *Image) String() string {
// 	return i.Path()
// }

func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}

	return temp.String()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	Delete(i *Image) error
	makeImagePath(galleryID uint) (string, error)
	ByGalleryID(galleryID uint) ([]Image, error)
}

type imageService struct {
}

func NewImageService() ImageService {
	return &imageService{}
}

// io.ReadCloser accepts anything ranging a mulitipart file to the string with a reader
// wrapped around it
func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {

	defer r.Close()

	// 1. Create a path for the image (e.g. images/galleries/20) via makeImagePath
	path, err := is.makeImagePath(galleryID)
	if err != nil {
		return err
	}

	// 2. Create a destination file according based on the filename passed in
	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}

	defer dst.Close()

	// 3. Copy the uploaded file data from the reader to the destination
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}

	return nil

}

func (is *imageService) makeImagePath(galleryID uint) (string, error) {
	galleryPath := is.imagePath(galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, err
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {

	path := is.imagePath(galleryID)

	imgStrings, err := filepath.Glob(path + "*")

	if err != nil {
		return nil, err
	}

	ret := make([]Image, len(imgStrings))
	for i := range imgStrings {
		imgStrings[i] = strings.Replace(imgStrings[i], path, "", 1)
		// strings[i] = "/" + strings[i]
		ret[i] = Image{
			Filename:  imgStrings[i],
			GalleryID: galleryID,
		}
	}

	return ret, nil
}

func (is *imageService) imagePath(galleryID uint) string {
	// Make a directory for the gallery image(s)
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}
