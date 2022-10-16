package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) Path() string {
	return "/" + i.RelativePath()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type ImageService interface {
	Create(ggallerID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (is *imageService) Create(gallerID uint, r io.ReadCloser, filename string) error {
	defer r.Close()
	path, err := is.mkImagePath(gallerID)
	if err != nil {
		return err
	}

	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imagePath(galleryID)
	imgPathes, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	images := make([]Image, len(imgPathes))
	for i := range imgPathes {
		imgPathes[i] = strings.Replace(imgPathes[i], path, "", 1)
		images[i] = Image{
			Filename:  imgPathes[i],
			GalleryID: galleryID,
		}
	}
	return images, nil
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := is.imagePath(galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}
