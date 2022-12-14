package models

import "github.com/jinzhu/gorm"

type Gallery struct {
	gorm.Model
	UserID uint    `gorm:"not_null;index"`
	Title  string  `gorm:"not_null"`
	Images []Image `gorm:"-"`
}

func (g *Gallery) ImagesSplitN(n int) [][]Image {
	imgLayout := make([][]Image, n)
	for i := 0; i < n; i++ {
		imgLayout[i] = make([]Image, 0)
	}
	for i, img := range g.Images {
		row := i % n
		imgLayout[row] = append(imgLayout[row], img)
	}
	return imgLayout
}

type GalleryService interface {
	GalleryDB
}

type GalleryDB interface {
	ByID(id uint) (*Gallery, error)
	ByUserID(userID uint) ([]Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error
}

type galleryService struct {
	GalleryDB
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{&galleryGorm{db}},
	}
}

type galleryValidator struct {
	GalleryDB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValidations(gallery,
		gv.titleRequired,
		gv.userIDRequired)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValidations(gallery,
		gv.titleRequired,
		gv.userIDRequired)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Update(gallery)
}

func (gv *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	err := runGalleryValidations(&gallery, gv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return gv.GalleryDB.Delete(id)
}

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)

	return &gallery, err
}

func (gg *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery
	err := gg.db.Where("user_id = ?", userID).Find(&galleries).Error
	if err != nil {
		return nil, err
	}
	return galleries, nil

}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&gallery).Error
}

type galleryValidationFunc func(*Gallery) error

func runGalleryValidations(gallery *Gallery, fns ...galleryValidationFunc) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

func (gv *galleryValidator) idGreaterThan(n uint) galleryValidationFunc {
	return galleryValidationFunc(func(gallery *Gallery) error {
		if gallery.ID <= n {
			return ErrInvalidId
		}
		return nil
	})
}
