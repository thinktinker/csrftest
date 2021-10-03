package models

import (
	"github.com/jinzhu/gorm"
)

// Gallery is our images container resource that visitors view
type Gallery struct {
	gorm.Model
	Title  string  `gorm:"not null"`
	UserID uint    `gorm:"not null;index"`
	Images []Image `gorm:"-"`
}

// GalleryDB interface exposes the methods that engages the database
// This interface is implemented by the GalleryService Interface
type GalleryDB interface {
	ByUserID(userID uint) ([]Gallery, error)
	ByID(id uint) (*Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(gallery *Gallery) error
}

// GalleryService interface implements GalleryBD
// so that it is ONLY able to consume methods that are exposed by GalleryDB
type GalleryService interface {
	GalleryDB
}

// galleryServices are internal services and objects that are not shared publicly
type galleryService struct {
	GalleryDB
}

// galleryValidator struct is a wrapper service around galleryService to perform validations
type galleryValidator struct {
	GalleryDB
}

// galleryGorm implement methods found in GalleryDB
type galleryGorm struct {
	db *gorm.DB
}

// this declaration validates that galleryGorm implements
// GalleryDB interface's methods
var _ GalleryDB = &galleryGorm{}

// this declaration checks that galleryValidator embeds GalleryDB interface's methods
var _ GalleryDB = &galleryValidator{}

// this declaration checks that galleryService embeds GalleryDB interface's methods
// At the same time, galleryService implements GalleryService struct's methods as well
var _ GalleryDB = &galleryService{}
var _ GalleryService = &galleryService{}

// ************** THIS SECTION CONTAINS THE METHODS TO MANAGE THE GALLERY IMAGES **************

func (g *Gallery) ImageSplitN(n int) [][]Image {
	ret := make([][]Image, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]Image, 0)
	}
	for i, img := range g.Images {
		// 0%3 = 0
		// 1%3 = 1
		// 2%3 = 2
		// 3%3 = 0
		// 4%3 = 1
		// 5%3 = 2
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}

	return ret
}

// ************** THIS SECTION CONTAINS THE GALLERYGORM METHODS FOR GALLERY **************

// NewGalleryService takes in a gorm.DB and return a pointer to the galleryService
func NewGalleryService(db *gorm.DB) GalleryService {
	// Note that what is returned to gs: the gorm.DB and the methods implemented by GalleryDB
	gs := &galleryGorm{db}

	return &galleryService{
		&galleryValidator{
			GalleryDB: gs,
		},
	}
}

// Create is a method implemented by galleryGorm struct
// which is part of GalleryDB interface's methods
func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

// Update is a method implemented by galleryGorm struct
// which is part of GalleryDB interface's methods
func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

// Delete will remove a gallery from the table
// IMPORTANT: Please make sure a value of > 0 is supplied, otherwise the entire table will be wiped out
// To ensure that the value is >0, implement the wrapper validator method idGreaterthan(n)
func (gg *galleryGorm) Delete(gallery *Gallery) error {
	return gg.db.Delete(&Gallery{Model: gorm.Model{ID: gallery.ID}}).Error
}

// ByID returns the gallery based on the parameter ID passed in
func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id=?", id)
	err := first(db, &gallery) //first function is declared in User's model; leave it there for now
	return &gallery, err
}

// ByUserID returns all the galleries that belongs to the parameter ID passed in
func (gg *galleryGorm) ByUserID(userId uint) ([]Gallery, error) {
	var galleries []Gallery
	err := gg.db.Where("user_id=?", userId).Find(&galleries).Error
	if err != nil {
		return nil, err
	}
	return galleries, nil
}

// ************** THIS SECTION CONTAINS THE VALIDATION CHAINING METHODS FOR GALLERY **************

// This user-defined function serves as the blueprint for all validation functions defined in this format
// eg. Create models after this user-defined function
type galleryValidateFunc func(*Gallery) error

// runGalleryValFunc function iterates through all the validations functions for the gallery
// As a variadic function it can contain accept zero validation functions as well
func runGalleryValFuncs(gallery *Gallery, fns ...galleryValidateFunc) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

// This Create validator runs before the passing to the galleryService method that creates the gallery
func (gv *galleryValidator) Create(gallery *Gallery) error {
	if err := runGalleryValFuncs(gallery,
		gv.userIDRequired,
		gv.titleRequired,
	); err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

// This Update validator runs before the passing to the galleryService method that updates the gallery
func (gv *galleryValidator) Update(gallery *Gallery) error {
	if err := runGalleryValFuncs(gallery,
		gv.userIDRequired,
		gv.titleRequired,
	); err != nil {
		return err
	}

	return gv.GalleryDB.Update(gallery)
}

// Delete will remove a gallery from the table
// IMPORTANT: Please make sure a value of > 0 is supplied, otherwise the entire table will be wiped out
func (gv *galleryValidator) Delete(gallery *Gallery) error {

	// This is included to prevent the entire table's record from being erased
	if err := runGalleryValFuncs(gallery, gv.idBeGreaterThan(0)); err != nil {
		return err
	}
	return gv.GalleryDB.Delete(gallery)
}

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserID <= 0 {
		return ErruserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

func (gv *galleryValidator) idBeGreaterThan(n uint) galleryValidateFunc {
	return galleryValidateFunc(func(gallery *Gallery) error {
		if gallery.ID <= n {
			return ErrInvalidID
		}
		return nil
	})
}
