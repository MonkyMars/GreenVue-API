package favorites

import (
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

// View name for PostgreSQL that fetches the user's favorites with listings and seller information
const viewName string = "user_favorites_with_listings_and_seller"

func GetFavorites(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.BadRequest("id is required")
	}

	client := db.NewSupabaseClient()

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	data, err := client.Query(viewName, "select=*&user_id=eq."+id)

	if err != nil {
		return errors.DatabaseError("Failed to fetch favorites: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.SuccessResponse(c, []db.FetchedListing{})
	}

	var favorites []db.FetchedListing
	if err = json.Unmarshal(data, &favorites); err != nil {
		return errors.InternalServerError("Failed to parse favorites data")
	}

	if favorites == nil {
		favorites = []db.FetchedListing{}
	}

	fmt.Println("favorites", favorites)

	return errors.SuccessResponse(c, favorites)
}

func AddFavorite(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	listingID := c.Params("listing_id")

	// Step 1: Validate params
	if userID == "" || listingID == "" {
		return errors.BadRequest("user_id and listing_id are required.")
	}

	client := db.NewSupabaseClient()
	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check server configuration.")
	}

	// Step 2: Check if favorite already exists
	query := fmt.Sprintf("select=*&user_id=eq.%s&listing_id=eq.%s", userID, listingID)
	existingFavorites, err := client.Query("favorites", query)
	if err != nil {
		return errors.DatabaseError("Failed to query favorites: " + err.Error())
	}
	if len(existingFavorites) > 0 {
		return errors.SuccessResponse(c, []lib.Favorite{})
	}

	// Step 3: Insert new favorite
	newFavorite := lib.Favorite{
		UserID:    userID,
		ListingID: listingID,
	}

	jsonBody, err := json.Marshal(newFavorite)
	if err != nil {
		return errors.InternalServerError("Failed to serialize favorite data.")
	}

	responseData, err := client.PostRaw("favorites", jsonBody)
	if err != nil {
		return errors.DatabaseError("Failed to insert favorite: " + err.Error())
	}

	// Step 4: Unmarshal response for returning to client
	var insertedFavorites []lib.Favorite
	if err := json.Unmarshal(responseData, &insertedFavorites); err != nil {
		return errors.InternalServerError("Failed to parse inserted favorite response: " + err.Error())
	}

	if len(insertedFavorites) == 0 {
		return errors.InternalServerError("Favorite insert returned empty response.")
	}

	// Step 5: Return created favorite
	return errors.SuccessResponse(c, insertedFavorites[0])
}
