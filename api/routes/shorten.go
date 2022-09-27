package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber"
	"github.com/google/uuid"
	"github.com/rinuccia/url-shortener/database"
	"github.com/rinuccia/url-shortener/helpers"
	"os"
	"strconv"
	"time"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	RateRemaining  int           `json:"rate_limit"`
	RateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) {

	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		return
	}

	r2 := database.NewClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		err = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
			return
		}
	} else {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
			return
		}
		if valInt <= 0 {
			limit, err := r2.TTL(database.Ctx, c.IP()).Result()
			if err != nil {
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
				return
			}
			c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
			return
		}
	}

	if !govalidator.IsURL(body.URL) {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid URL"})
		return
	}

	if !helpers.RemoveDomainError(body.URL) {
		c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "you can't hack the system(:"})
		return
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.NewClient(0)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "URL custom short is already in use"})
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*time.Hour).Err()
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to server"})
		return
	}

	resp := response{
		URL:            body.URL,
		Expiry:         body.Expiry,
		RateRemaining:  10,
		RateLimitReset: 30,
	}

	err = r2.Decr(database.Ctx, c.IP()).Err()
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.RateRemaining, err = strconv.Atoi(val)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}

	ttl, err := r2.TTL(database.Ctx, c.IP()).Result()
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}
	resp.RateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	c.Status(fiber.StatusOK).JSON(resp)
	return
}
