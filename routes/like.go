package routes

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iamsaidovibra/blog-rest-api/database"
	"github.com/iamsaidovibra/blog-rest-api/models"
	"github.com/iamsaidovibra/blog-rest-api/utils"
	"gorm.io/gorm"
)

type LikeSerializer struct {
	ID        uint              `json:"id"`
	User      UserSerializer    `json:"user"`
	Article   ArticleSerializer `json:"article"`
	CreatedAt string            `json:"liked_at"`
}

func CreateLike(c *fiber.Ctx) error {
	articleID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Article ID must be an integer"})
	}

	userID := utils.GetUserID(c)

	var article models.Article
	err = database.Database.Db.First(&article, articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	like := models.Like{UserID: userID, ArticleID: uint(articleID)}
	err = database.Database.Db.Create(&like).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return c.Status(400).JSON(fiber.Map{"error": "Already liked"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Could not create like"})
	}

	database.Database.Db.Preload("User").Preload("Article.Author").First(&like, like.ID)

	return c.Status(201).JSON(
		LikeSerializer{
			ID:        like.ID,
			User:      CreateResponseUser(like.User),
			Article:   CreateResponseArticle(like.Article, CreateResponseUser(like.Article.Author)),
			CreatedAt: like.CreatedAt.Format(time.RFC3339),
		},
	)
}

func DeleteLike(c *fiber.Ctx) error {
	articleID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Article ID must be an integer"})
	}

	userID := utils.GetUserID(c)
	var like models.Like
	err = database.Database.Db.Where("user_id = ? AND article_id = ?", userID, articleID).
		First(&like).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "Like not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	database.Database.Db.Delete(&like)
	return c.SendStatus(204)
}

func GetMyLikes(c *fiber.Ctx) error {
	userID := utils.GetUserID(c)
	var likes []models.Like
	err := database.Database.Db.
		Where("user_id = ?", userID).
		Preload("Article.Author").
		Find(&likes).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch likes"})
	}

	// serialize
	response := make([]LikeSerializer, len(likes))
	for i, l := range likes {
		response[i] = LikeSerializer{
			ID:        l.ID,
			User:      CreateResponseUser(models.User{Model: gorm.Model{ID: userID}}),
			Article:   CreateResponseArticle(l.Article, CreateResponseUser(l.Article.Author)),
			CreatedAt: l.CreatedAt.Format(time.RFC3339),
		}
	}
	return c.Status(200).JSON(response)
}
