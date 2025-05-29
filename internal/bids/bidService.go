package bids

import (
	"encoding/json"
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/validation"
	"sort"
)

// BidService handles bid-related business logic
type BidService struct {
	client *db.SupabaseClient
}

// NewBidService creates a new bid service
func NewBidService() *BidService {
	return &BidService{
		client: db.GetGlobalClient(),
	}
}

// GetListingWithBids retrieves a listing with its current bids
func (bs *BidService) GetListingWithBids(listingID string) (*lib.FetchedListing, []lib.FetchedBid, error) {
	if bs.client == nil {
		return nil, nil, fmt.Errorf("database client not available")
	}
	// Get listing details
	listingQuery := fmt.Sprintf("id=eq.%s", listingID)
	listingData, err := bs.client.GET("listing_details", listingQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve listing: %w", err)
	}

	var listings []lib.FetchedListing
	if err := json.Unmarshal(listingData, &listings); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal listing data: %w", err)
	}

	if len(listings) == 0 {
		return nil, nil, fmt.Errorf("listing not found")
	}

	listing := &listings[0]

	// Get bids for this listing
	bidQuery := fmt.Sprintf("listing_id=eq.%s&order=price.desc,created_at.desc", listingID)
	bidData, err := bs.client.GET("fetched_bids", bidQuery)
	if err != nil {
		return listing, nil, fmt.Errorf("failed to retrieve bids: %w", err)
	}

	var bids []lib.FetchedBid
	if err := json.Unmarshal(bidData, &bids); err != nil {
		return listing, nil, fmt.Errorf("failed to unmarshal bids data: %w", err)
	}

	return listing, bids, nil
}

// GetHighestBid returns the highest bid for a listing
func (bs *BidService) GetHighestBid(bids []lib.FetchedBid) *lib.FetchedBid {
	if len(bids) == 0 {
		return nil
	}

	// Sort bids by price (descending) and created_at (descending for ties)
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].Price == bids[j].Price {
			return bids[i].CreatedAt.After(bids[j].CreatedAt)
		}
		return bids[i].Price > bids[j].Price
	})

	return &bids[0]
}

// ValidateBidContext creates validation context for a bid
func (bs *BidService) ValidateBidContext(bid lib.Bid) (*validation.BidValidationContext, error) {
	listing, bids, err := bs.GetListingWithBids(bid.ListingID.String())
	if err != nil {
		return nil, err
	}

	highestBid := bs.GetHighestBid(bids)

	return &validation.BidValidationContext{
		Listing:      listing,
		HighestBid:   highestBid,
		ExistingBids: bids,
		BidderID:     bid.UserID,
	}, nil
}

// PlaceBid handles the complete bid placement process
func (bs *BidService) PlaceBid(bid lib.Bid) (*lib.FetchedBid, error) {
	if bs.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Create validation context
	context, err := bs.ValidateBidContext(bid)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation context: %w", err)
	}

	// Validate the bid
	validationResult := validation.ValidateBid(bid, context)
	if !validationResult.Valid {
		return nil, fmt.Errorf("bid validation failed: %v", validationResult.Errors)
	}

	// Place the bid in database
	bidData, err := bs.client.POST("bids", map[string]any{
		"listing_id": bid.ListingID,
		"user_id":    bid.UserID,
		"bid":        bid.Price,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store bid: %w", err)
	}

	// Return the newly created bid with user info
	var createdBids []map[string]any
	if err := json.Unmarshal(bidData, &createdBids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created bid: %w", err)
	}

	if len(createdBids) == 0 {
		return nil, fmt.Errorf("no bid was created")
	}

	// Get the full bid details with user info
	bidID, ok := createdBids[0]["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid bid ID returned")
	}

	query := fmt.Sprintf("id=eq.%s", bidID)
	fetchedBidData, err := bs.client.GET("fetched_bids", query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created bid: %w", err)
	}

	var fetchedBids []lib.FetchedBid
	if err := json.Unmarshal(fetchedBidData, &fetchedBids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fetched bid: %w", err)
	}

	if len(fetchedBids) == 0 {
		return nil, fmt.Errorf("created bid not found")
	}

	return &fetchedBids[0], nil
}
