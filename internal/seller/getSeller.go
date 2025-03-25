package seller

import (
	"greentrade-eu/internal/db"
	"github.com/gofiber/fiber/v2"
	"encoding/json"
	"fmt"
)

func GetSellers(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	data, err := client.Query("sellers", "select=*")

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch sellers: " + err.Error(),
		})
	}

	var sellers []db.Seller
	if err := json.Unmarshal(data, &sellers); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse sellers: " + err.Error(),
		})
	}

	return c.JSON(sellers)
};

func GetSellerById(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	sellerID := c.Params("id")
	query := fmt.Sprintf("select=*&id=eq.%s", sellerID)
	data, err := client.Query("sellers", query)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch seller: " + err.Error(),
		})
	}

	var sellers []db.Seller
	if err := json.Unmarshal(data, &sellers); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse seller: " + err.Error(),
		})
	}

	if len(sellers) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Seller not found",
		})
	}

	return c.JSON(sellers[0])
};