package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	Content   string  `json:"content" gorm:"type:text;not null"`
	UserID    uint    `json:"user_id" gorm:"not null"`
	ArticleID uint    `json:"article_id" gorm:"not null"`
	User      User    `json:"user" gorm:"foreignKey:UserID"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}

