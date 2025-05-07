package utils

import "github.com/gofiber/fiber/v2"

func Paginate(c *fiber.Ctx) (limit int, offset int) {
	limit = c.QueryInt("limit", 10)
	offset = c.QueryInt("offset", 0)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return
}
