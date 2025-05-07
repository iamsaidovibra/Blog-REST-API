package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/iamsaidovibra/blog-rest-api/database"
	"github.com/iamsaidovibra/blog-rest-api/routes"
	"github.com/iamsaidovibra/blog-rest-api/utils"
)

func welcome(c *fiber.Ctx) error {
	return c.SendString("hi there")
}


func setupRoutes(app fiber.Router) {
	// the “welcome” now lives at GET /api/
	app.Get("/", welcome)

	// users:
	app.Get("/users", routes.GetUsers) // GET  /api/users
	app.Get("/users/:id", routes.GetUserById)
	app.Put("/users/:id", routes.UpdateUser)
	app.Delete("/users/:id", routes.DeleteUser)

	// articles:
	app.Post("/article", routes.CreateArticle)
	app.Get("/article", routes.GetArticles)
	app.Get("/article/:id", routes.GetArticleById)
	app.Put("/article/:id", routes.UpdateArticle)
	app.Delete("/article/:id", routes.DeleteArticle)
	app.Get("/search", routes.SearchArticles)


	// likes:
	app.Post("/likes", routes.CreateLike)
	app.Delete("/like/:id", routes.DeleteLike)
	app.Get("/like/:id", routes.GetMyLikes)

	//comments:
	app.Post("/comments/:id", routes.CreateComment)
	app.Put("/comments/:id", routes.UpdateComment)
	app.Delete("/comments/:id", routes.DeleteComment)

}

func main() {
	database.ConnectDb()
	app := fiber.New()

	// Public routes (no authentication required)
	app.Post("/login", routes.LoginUser)
	app.Post("/users", routes.CreateUser) 
	app.Get("/search", routes.SearchArticles)

	// Protected routes (require JWT)
	protected := app.Group("/api", utils.Protect)
	setupRoutes(protected)

	log.Fatal(app.Listen(":3000"))
}


// API для блога (аналог Medium)
// CRUD для статей и комментариев.
// Пагинация и поиск.
// Доп. задание: добавить лайки.
