package bids

import (
	"encoding/json"
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"sort"

	"github.com/gofiber/fiber/v2"
)

// GetBids retrieves all bids from the database from a specific listing with enhanced sorting
func GetBids(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to get database client")
	}

	listingID := c.Params("listing_id")
	if listingID == "" {
		return errors.BadRequest("Listing ID is required")
	}

	// Get sorting preference from query params (default: price descending)
	sortBy := c.Query("sort", "price")
	order := c.Query("order", "desc")

	query := fmt.Sprintf("listing_id=eq.%s", listingID)
	data, err := client.GET("fetched_bids", query)

	if err != nil {
		return errors.InternalServerError("Failed to retrieve bids: " + err.Error())
	}

	var bids []lib.FetchedBid
	if err := json.Unmarshal(data, &bids); err != nil {
		return errors.InternalServerError("Failed to unmarshal bids data: " + err.Error())
	}

	// Sort bids based on query parameters
	sortBids(bids, sortBy, order)

	return errors.SuccessResponse(c, bids)
}

// sortBids sorts the bids slice based on the given criteria
func sortBids(bids []lib.FetchedBid, sortBy, order string) {
	switch sortBy {
	case "price":
		if order == "desc" {
			sort.Slice(bids, func(i, j int) bool {
				if bids[i].Price == bids[j].Price {
					return bids[i].CreatedAt.After(bids[j].CreatedAt)
				}
				return bids[i].Price > bids[j].Price
			})
		} else {
			sort.Slice(bids, func(i, j int) bool {
				if bids[i].Price == bids[j].Price {
					return bids[i].CreatedAt.Before(bids[j].CreatedAt)
				}
				return bids[i].Price < bids[j].Price
			})
		}
	case "time":
		if order == "desc" {
			sort.Slice(bids, func(i, j int) bool {
				return bids[i].CreatedAt.After(bids[j].CreatedAt)
			})
		} else {
			sort.Slice(bids, func(i, j int) bool {
				return bids[i].CreatedAt.Before(bids[j].CreatedAt)
			})
		}
	case "user":
		if order == "desc" {
			sort.Slice(bids, func(i, j int) bool {
				return bids[i].UserName > bids[j].UserName
			})
		} else {
			sort.Slice(bids, func(i, j int) bool {
				return bids[i].UserName < bids[j].UserName
			})
		}
	default:
		// Default to price descending
		sort.Slice(bids, func(i, j int) bool {
			if bids[i].Price == bids[j].Price {
				return bids[i].CreatedAt.After(bids[j].CreatedAt)
			}
			return bids[i].Price > bids[j].Price
		})
	}
}
