package seller

/*
	id uuid not null,
  	created_at timestamp with time zone not null default now(),
  	description text null,
  	rating integer not null default 0,
  	verified boolean not null default false,
*/

import(
	"greentrade-eu/internal/db"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func CreateSeller(c *fiber.Ctx) error {
	// rating and verified are set to default values. No need to provide them.

	// Create a new Supabase client
	client := db.NewSupabaseClient()

	// Define payload struct

	var payload struct {
		Description string `json:"description"`
		ID string `json:"id"`
	}

	// Parse JSON request body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON payload: " + err.Error(),
		})
	}

	Description := payload.Description
	ID := payload.ID

	// Validate required fields
	if Description == "" || ID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Description and ID are required",
		})
	}

	// Insert seller into the database
	err := client.InsertSeller(ID, Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to store seller in database: " + err.Error(),
		})
	}

	response := fmt.Sprintf("Seller with ID %s created successfully", ID)

	// Return success response
	return c.Status(201).JSON(fiber.Map{
		"message": response,
		"sellerId": ID,
	})
}
