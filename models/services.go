package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DBConnectionServices unifies all the connections to the database
type Services struct {
	Gallery GalleryService
	User    UserService
	Image   ImageService
	db      *gorm.DB //both NewUserService and the methods here are accessing the same reference of gorm.DB
}

type ServicesConfig func(*Services) error

func NewServices(cfgs ...ServicesConfig) (*Services, error) {

	var s Services

	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

// Close the database connection
func (s *Services) Close() error {
	return s.db.Close()
}

// Destructive Reset allows the requestor the drop the existing database tables and re-create them for testing
// NOT for production use
func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&User{}, &Gallery{}).Error
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

// Automigrate will attempt to automatically migrate the users table
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{}).Error
}

// func AddImageService(services *DBServices) error {
// 	services.Image = NewImageService()
// 	return nil
// }

func WithGorm(dialect, connectionstring string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionstring)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

// For the With...() functions below,
// it is the closure here that matches ServicesConfig user-defined function
// which NewDBServices accepts as its parameter(s)

func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db, pepper, hmacKey)
		return nil
	}
}

func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}

func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService()
		return nil
	}
}

func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)
		return nil
	}
}

// New Way
// func NewDBServices(cfg ...func(*DBServices) error) (*DBServices, error) {
// Original way of using NewDBServices
// func NewDBServices(dialect, connectionstring string) (*DBServices, error) {

// 	// TO DO: CONFIG THIS
// 	db, err := gorm.Open(dialect, connectionstring)
// 	if err != nil {
// 		return nil, err
// 	}

// 	db.LogMode(true)

// 	return &DBServices{
// 		// Gallery services shall be updated later
// 		User:    NewUserService(db),
// 		Gallery: NewGalleryService(db),
// 		Image:   NewImageService(),
// 		db:      db,
// 	}, nil
// }
