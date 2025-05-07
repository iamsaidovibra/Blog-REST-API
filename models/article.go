package models

import "gorm.io/gorm"

type Article struct {
	gorm.Model
	Title    string    `json:"title" gorm:"not null"`
	Content  string    `json:"content" gorm:"type:text;not null"`
	AuthorID uint      `json:"author_id" gorm:"not null"`
	Author   User      `json:"author" gorm:"foreignKey:AuthorID"`
	Comments []Comment `json:"comments" gorm:"foreignKey:ArticleID"`
	Likes    []Like    `json:"likes" gorm:"foreignKey:ArticleID"`
}

