package routes

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iamsaidovibra/blog-rest-api/database"
	"github.com/iamsaidovibra/blog-rest-api/models"
	"github.com/iamsaidovibra/blog-rest-api/utils"
	"gorm.io/gorm"
)

// CommentSerializer shapes the JSON response for a comment
type CommentSerializer struct {
	ID          uint              `json:"id"`
	Content     string            `json:"content"`
	User        UserSerializer    `json:"user"`
	Article     ArticleSerializer `json:"article"`
	CommentedAt time.Time         `json:"commented_at"`
}

// CreateCommentInput defines what clients can send when creating or updating a comment
type CreateCommentInput struct {
	Content string `json:"content" validate:"required"`
}

// CreateComment handles POST /api/articles/:id/comment
func CreateComment(c *fiber.Ctx) error {
	// 1) extract article ID from URL
	articleID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Article ID must be an integer"})
	}

	// 2) parse content
	var input CreateCommentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// 3) get authenticated user ID
	userID := utils.GetUserID(c)

	// 4) ensure article exists
	var article models.Article
	err = database.Database.Db.First(&article, articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// 5) create comment
	comment := models.Comment{
		Content:   input.Content,
		UserID:    userID,
		ArticleID: uint(articleID),
	}
	err = database.Database.Db.Create(&comment).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create comment"})
	}

	// 6) preload associations
	database.Database.Db.Preload("User").Preload("Article.Author").First(&comment, comment.ID)

	// 7) respond
	return c.Status(201).JSON(
		CommentSerializer{
			ID:          comment.ID,
			Content:     comment.Content,
			User:        CreateResponseUser(comment.User),
			Article:     CreateResponseArticle(comment.Article, CreateResponseUser(comment.Article.Author)),
			CommentedAt: comment.CreatedAt,
		},
	)
}

// UpdateComment handles PUT /api/comments/:id
func UpdateComment(c *fiber.Ctx) error {
	// 1) extract comment ID
	commentID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Comment ID must be an integer"})
	}

	// 2) get authenticated user ID
	userID := utils.GetUserID(c)

	// 3) find comment belonging to user
	var comment models.Comment
	err = database.Database.Db.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "Comment not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// 4) parse update data
	var input CreateCommentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// 5) update fields
	comment.Content = input.Content

	// 6) save changes
	if err := database.Database.Db.Save(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update comment"})
	}

	// 7) preload associations
	database.Database.Db.Preload("User").Preload("Article.Author").First(&comment, comment.ID)

	// 8) respond
	return c.Status(200).JSON(
		CommentSerializer{
			ID:          comment.ID,
			Content:     comment.Content,
			User:        CreateResponseUser(comment.User),
			Article:     CreateResponseArticle(comment.Article, CreateResponseUser(comment.Article.Author)),
			CommentedAt: comment.UpdatedAt,
		},
	)
}

// DeleteComment handles DELETE /api/comments/:id
func DeleteComment(c *fiber.Ctx) error {
	// 1) extract comment ID
	commentID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Comment ID must be an integer"})
	}

	// 2) get authenticated user ID
	userID := utils.GetUserID(c)

	// 3) find comment belonging to user
	var comment models.Comment
	err = database.Database.Db.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "Comment not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// 4) delete it
	database.Database.Db.Delete(&comment)
	return c.SendStatus(204)
}

// GetMyComments handles GET /api/comments
func GetMyComments(c *fiber.Ctx) error {
	// 1) get authenticated user ID
	userID := utils.GetUserID(c)

	// 2) load user's comments
	var comments []models.Comment
	err := database.Database.Db.
		Where("user_id = ?", userID).
		Preload("Article.Author").
		Find(&comments).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch comments"})
	}

	// 3) serialize
	response := make([]CommentSerializer, len(comments))
	for i, cm := range comments {
		response[i] = CommentSerializer{
			ID:          cm.ID,
			Content:     cm.Content,
			User:        CreateResponseUser(models.User{Model: gorm.Model{ID: userID}}),
			Article:     CreateResponseArticle(cm.Article, CreateResponseUser(cm.Article.Author)),
			CommentedAt: cm.CreatedAt,
		}
	}

	return c.Status(200).JSON(response)
}
