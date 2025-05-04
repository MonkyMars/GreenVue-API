package db

import (
	"context"
	"fmt"
	"greentrade-eu/lib"

	"github.com/gofiber/fiber/v2"
)

// Repository defines standard CRUD operations for all entities
type Repository interface {
	Get(ctx context.Context, params lib.QueryParams) ([]byte, error)
	GetByID(ctx context.Context, table string, id string) ([]byte, error)
	Create(ctx context.Context, table string, data any) ([]byte, error)
	Update(ctx context.Context, table string, id string, data any) ([]byte, error)
	Delete(ctx context.Context, table string, id string) error
}

// SupabaseRepository implements the Repository interface with Supabase
type SupabaseRepository struct {
	client *SupabaseClient
}

// NewSupabaseRepository creates a new Supabase repository
func NewSupabaseRepository(client *SupabaseClient) *SupabaseRepository {
	return &SupabaseRepository{
		client: client,
	}
}

// Get fetches records based on query parameters
func (r *SupabaseRepository) Get(ctx context.Context, params lib.QueryParams) ([]byte, error) {
	query := ""

	if params.Filter != "" {
		query += params.Filter
	}

	if params.Limit > 0 {
		if query != "" {
			query += "&"
		}
		query += "limit=" + fmt.Sprint(params.Limit)
	}

	if params.Offset > 0 {
		if query != "" {
			query += "&"
		}
		query += "offset=" + fmt.Sprint(params.Offset)
	}

	if params.OrderBy != "" {
		if query != "" {
			query += "&"
		}
		orderDirection := "asc"
		if params.Direction != "" {
			orderDirection = params.Direction
		}
		query += "order=" + params.OrderBy + "." + orderDirection
	}

	return r.client.GET(&fiber.Ctx{}, params.Table, query)
}

// GetByID fetches a record by ID
func (r *SupabaseRepository) GetByID(ctx context.Context, table string, id string) ([]byte, error) {
	query := "id=eq." + id
	return r.client.GET(&fiber.Ctx{}, table, query)
}

// Create creates a new record
func (r *SupabaseRepository) Create(ctx context.Context, table string, data any) ([]byte, error) {
	return r.client.POST(&fiber.Ctx{}, table, data)
}

// Update updates a record by ID
func (r *SupabaseRepository) Update(ctx context.Context, table string, id string, data any) ([]byte, error) {
	return r.client.PATCH(&fiber.Ctx{}, table, id, data)
}

// Delete deletes a record by ID
func (r *SupabaseRepository) Delete(ctx context.Context, table string, id string) error {
	_, err := r.client.DELETE(&fiber.Ctx{}, table, "id=eq."+id)
	return err
}

// Special operations that don't fit the standard CRUD pattern

// UploadImage uploads an image to Supabase storage
func (r *SupabaseRepository) UploadImage(ctx context.Context, filename, bucket string, image []byte) (string, error) {
	_, err := r.client.UploadImage(filename, bucket, image)
	if err != nil {
		return "", err
	}

	// Construct and return the public URL
	imageURL := r.client.URL + "/storage/v1/object/public/" + bucket + "/" + filename
	return imageURL, nil
}

// Auth operations

// SignUp registers a new user
func (r *SupabaseRepository) SignUp(ctx context.Context, email, password string) (*lib.User, error) {
	return r.client.SignUp(email, password)
}

// Login authenticates a user
func (r *SupabaseRepository) Login(ctx context.Context, email, password string) (*lib.AuthResponse, error) {
	return r.client.Login(email, password)
}
