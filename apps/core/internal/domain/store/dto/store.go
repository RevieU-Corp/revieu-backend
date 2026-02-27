package dto

// CreateStoreRequest is the request payload for creating a store.
type CreateStoreRequest struct {
	Name          string   `json:"name" binding:"max=255"`
	Description   string   `json:"description" binding:"max=2000"`
	Address       string   `json:"address" binding:"max=255"`
	City          string   `json:"city" binding:"max=100"`
	State         string   `json:"state" binding:"max=100"`
	ZipCode       string   `json:"zip_code" binding:"max=20"`
	Country       string   `json:"country" binding:"max=50"`
	Phone         string   `json:"phone" binding:"max=50"`
	Website       string   `json:"website" binding:"max=255"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	CoverImageURL string   `json:"cover_image_url" binding:"max=255"`
	Images        []string `json:"images"`
}
