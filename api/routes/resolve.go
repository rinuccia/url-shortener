package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber"
	"github.com/rinuccia/url-shortener/database"
)

func ResolveURL(c *fiber.Ctx) {
	url := c.Params("url")

	r := database.NewClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found in database"})
		return
	} else if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to DB"})
		return
	}

	rInr := database.NewClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")

	c.Redirect(value, fiber.StatusMovedPermanently)
	return
}
