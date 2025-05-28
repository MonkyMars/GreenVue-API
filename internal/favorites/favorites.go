package favorites

import (
	"fmt"
	"greenvue/internal/auth"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"log"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// View name for PostgreSQL that fetches the user's favorites with listings and seller information
const viewName string = "user_favorites"

func GetFavorites(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*auth.Claims)

	if !ok {
		log.Println(claims, ok)
		return errors.Unauthorized("Invalid or missing authentication")
	}

	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	data, err := client.GET(viewName, "select=*&user_id=eq."+claims.UserId.String())

	if err != nil {
		return errors.DatabaseError("Failed to fetch favorites: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.SuccessResponse(c, []lib.FetchedFavorite{})
	}

	var favorites []lib.FetchedFavorite
	if err = json.Unmarshal(data, &favorites); err != nil {
		return errors.InternalServerError("Failed to parse favorites data")
	}

	if favorites == nil {
		favorites = []lib.FetchedFavorite{}
	}

	return errors.SuccessResponse(c, favorites)
}

func AddFavorite(c *fiber.Ctx) error {

	var payload struct {
		ListingID uuid.UUID `json:"listing_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request body: " + err.Error())
	}

	claims, ok := c.Locals("user").(*auth.Claims)
	if !ok {
		return errors.Unauthorized("Invalid or missing authentication")
	}

	// Step 1: Validate params
	if payload.ListingID == uuid.Nil {
		return errors.BadRequest("listing_id is required.")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check server configuration.")
	}

	// Step 2: Check if favorite already exists
	query := fmt.Sprintf("select=*&user_id=eq.%s&listing_id=eq.%s", claims.UserId, payload.ListingID)
	existingFavorites, err := client.GET("favorites", query)
	if err != nil {
		return errors.DatabaseError("Failed to query favorites: " + err.Error())
	}

	if string(existingFavorites) != "[]" {
		return errors.BadRequest("Favorite already exists.")
	}

	// Step 3: Create new favorite using standardized POST operation
	newFavorite := lib.Favorite{
		UserID:    claims.UserId,
		ListingID: payload.ListingID,
	}

	responseData, err := client.POST("favorites", newFavorite)
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

func DeleteFavorite(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*auth.Claims)
	if !ok {
		return errors.Unauthorized("Invalid or missing authentication")
	}

	userID := claims.UserId

	listingID := c.Params("listing_id")

	// Step 1: Validate params
	if listingID == "" {
		return errors.BadRequest("listing_id is required.")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check server configuration.")
	}

	// Step 2: Delete favorite using the standardized DELETE operation
	query := fmt.Sprintf("user_id=eq.%s&listing_id=eq.%s", userID, listingID)
	responseData, err := client.DELETE("favorites", query)
	if err != nil {
		return errors.DatabaseError("Failed to delete favorite: " + err.Error())
	}

	if string(responseData) == "[]" {
		return errors.SuccessResponse(c, []lib.Favorite{})
	}

	return errors.SuccessResponse(c, responseData)
}

func IsFavorite(c *fiber.Ctx) error {
	listingID := c.Params("listing_id")

	claims, ok := c.Locals("user").(*auth.Claims)

	if !ok {
		return errors.Unauthorized("Invalid or missing authentication")
	}

	userID := claims.UserId.String()
	// Step 1: Validate params
	if userID == "" || listingID == "" {
		return errors.BadRequest("user_id and listing_id are required.")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check server configuration.")
	}

	// Step 2: Check if favorite exists
	query := fmt.Sprintf("user_id=eq.%s&listing_id=eq.%s", userID, listingID)
	existingFavorites, err := client.GET("favorites", query)
	if err != nil {
		return errors.DatabaseError("Failed to query favorites: " + err.Error())
	}

	if string(existingFavorites) == "[]" {
		return errors.SuccessResponse(c, false)
	}

	return errors.SuccessResponse(c, true)
}
