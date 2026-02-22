package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"                  json:"id"`
	Name          string         `gorm:"type:varchar(100);not null"            json:"name"`
	Email         string         `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"type:varchar(255);not null"            json:"-"` // never serialised
	EmailVerified bool           `gorm:"type:boolean;default:false" json:"-"`
	CreatedAt     time.Time      `                                             json:"created_at"`
	UpdatedAt     time.Time      `                                             json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                                 json:"-"` // soft-delete
}

// BeforeCreate assigns a new UUID before inserting a record.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
