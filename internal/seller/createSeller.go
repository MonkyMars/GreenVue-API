package seller

/*
	id uuid not null,
  	created_at timestamp with time zone not null default now(),
  	description text null,
  	rating integer not null default 0,
  	verified boolean not null default false,
*/

import (
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func CreateSeller(c *fiber.Ctx) error {
	// rating and verified are set to default values. No need to provide them.

	// Create a new Supabase client
	client := db.NewSupabaseClient()

	// Define payload struct

	var payload struct {
		Description string `json:"description"`
		ID          string `json:"id"`
	}

	// Parse JSON request body
	if err := c.BodyParser(&payload); err != nil {
		errors.BadRequest("Failed to parse request body: " + err.Error())
	}

	Description := payload.Description
	ID := payload.ID

	// Validate required fields
	if Description == "" || ID == "" {
		return errors.BadRequest("Description and ID are required fields")
	}

	// Insert seller into the database
	err := client.InsertSeller(ID, Description)
	if err != nil {
		return errors.InternalServerError("Failed to create seller: " + err.Error())
	}

	response := fmt.Sprintf("Seller with ID %s created successfully", ID)

	// Return success response
	return errors.SuccessResponse(c, response)
}
