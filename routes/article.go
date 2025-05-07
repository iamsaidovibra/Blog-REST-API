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

type ArticleSerializer struct {
	ID        uint           `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Author    UserSerializer `json:"author"`
	CreatedAt time.Time      `json:"publication_date"`
	Likes     uint 			`json:"likes"`
	Comments  []models.Comment `json:"comments"`
}

type CreateArticleInput struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
}

// type GetUserInput struct {
// }

func CreateResponseArticle(article models.Article, author UserSerializer) ArticleSerializer {
	return ArticleSerializer{
		ID:        article.ID,
		Title:     article.Title,
		Content:   article.Content,
		Author:    author,
		CreatedAt: article.CreatedAt,
	}
}


func CreateArticle(c *fiber.Ctx) error {
	// 1) parse only title & content
	var input CreateArticleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	userID := utils.GetUserID(c)

	article := models.Article{
		Title:    input.Title,
		Content:  input.Content,
		AuthorID: userID,
	}

	// 4) save to the DB
	if err := database.Database.Db.Create(&article).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create article"})
	}

	// 5) preload the Author so your serializer can use it
	database.Database.Db.Preload("Author").First(&article, article.ID)

	// 6) respond
	return c.Status(201).JSON(
		CreateResponseArticle(article, CreateResponseUser(article.Author)),
	)
}


func findArticle(id uint, userId uint, article *models.Article) error {
	err := database.Database.Db.
		Where("id = ? AND author_id = ?", id, userId).
		First(article).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Article not found or access forbidden")
		}
		return err
	}
	return nil
}

func GetArticles(c *fiber.Ctx) error {
	userID := utils.GetUserID(c)
	limit, offset := utils.Paginate(c)

	var articles []models.Article
	if err := database.Database.Db.
		Where("author_id = ?", userID).
		Preload("Author").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch articles"})
	}

	// 3) serialize
	response := make([]ArticleSerializer, len(articles))
	for i, art := range articles {
		response[i] = CreateResponseArticle(art, CreateResponseUser(art.Author))
	}

	return c.Status(200).JSON(response)
}

func GetArticleById(c *fiber.Ctx) error {

	userID := utils.GetUserID(c)
	var article models.Article
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON("Make sure Id is an integer")
	}

	if err := findArticle(uint(id), uint(userID), &article); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	responseUser := CreateResponseUser(article.Author)
	responseArticle := CreateResponseArticle(article, responseUser)
	return c.Status(201).JSON(responseArticle)
}

func UpdateArticle(c *fiber.Ctx) error {
	userId := utils.GetUserID(c)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON("Invalid article ID")
	}

	var article models.Article
	if err := findArticle(uint(id), uint(userId), &article); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
	}

	type UpdateArticle struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var updateData UpdateArticle
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update fields
	article.Title = updateData.Title
	article.Content = updateData.Content

	if err := database.Database.Db.Save(&article).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save changes"})
	}

	var author models.User
	if err := database.Database.Db.First(&author, article.AuthorID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load author"})
	}

	responseUser := CreateResponseUser(author)
	responseArticle := CreateResponseArticle(article, responseUser)

	return c.Status(200).JSON(responseArticle)
}

func SearchArticles(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Query parameter 'q' is required"})
	}

	limit, offset := utils.Paginate(c)
	var articles []models.Article
	if err := database.Database.Db.
		Where("LOWER(title) LIKE ?", "%"+query+"%").
		Preload("Author").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to search articles"})
	}

	response := make([]ArticleSerializer, len(articles))
	for i, art := range articles {
		response[i] = CreateResponseArticle(art, CreateResponseUser(art.Author))
	}

	return c.Status(200).JSON(response)
}

func GetCommentsForArticle(c *fiber.Ctx) error {
	 articleID, err := c.ParamsInt("id")
	 if err != nil {
	  return c.Status(400).JSON(fiber.Map{"error": "Article ID must be an integer"})
	 }
	
	 var comments []models.Comment
	 err = database.Database.Db.
	  Where("article_id = ?", articleID).
	  Preload("User").           
	  Preload("Article.Author"). 
	  Find(&comments).Error
	 if err != nil {
	  return c.Status(500).JSON(fiber.Map{"error": "Could not fetch comments"})
	 }
	
	 response := make([]CommentSerializer, len(comments))
	 for i, cm := range comments {
	  response[i] = CommentSerializer{
	   ID:          cm.ID,
	   Content:     cm.Content,
	   User:        CreateResponseUser(cm.User), 
	   Article:     CreateResponseArticle(cm.Article, CreateResponseUser(cm.Article.Author)),
	   CommentedAt: cm.CreatedAt,
	  }
	 }
		return c.Status(200).JSON(response)
	}

func DeleteArticle(c *fiber.Ctx) error {
	userId := utils.GetUserID(c)
	id, err := c.ParamsInt("id")
	var article models.Article

	if err != nil {
		return c.Status(400).JSON("Make sure ID is an integer")
	}

	if err := findArticle(uint(id), uint(userId), &article); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	if err := database.Database.Db.Delete(&article).Error; err != nil {
		return c.Status(404).JSON(err.Error())
	}

	return c.Status(200).SendString("Article was DELETED successfully")
}



// func GetArticleForUser(c *fiber.Ctx)error{
// 	id, err :=
// }

//limit and offset for pagination
