package models

import "gorm.io/gorm"

type Like struct {
	gorm.Model
	UserID    uint    `json:"user_id" gorm:"not null;uniqueIndex:idx_user_article"`
	ArticleID uint    `json:"article_id" gorm:"not null;uniqueIndex:idx_user_article"`
	User      User    `json:"user" gorm:"foreignKey:UserID"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}
