package dto

// CreateStoreRequest is the request payload for creating a store.
type CreateStoreRequest struct {
	Name          string             `json:"name" binding:"max=255"`
	Description   string             `json:"description" binding:"max=2000"`
	Address       string             `json:"address" binding:"max=255"`
	City          string             `json:"city" binding:"max=100"`
	State         string             `json:"state" binding:"max=100"`
	ZipCode       string             `json:"zip_code" binding:"max=20"`
	Country       string             `json:"country" binding:"max=50"`
	Phone         string             `json:"phone" binding:"max=50"`
	Website       string             `json:"website" binding:"max=255"`
	Latitude      float64            `json:"latitude"`
	Longitude     float64            `json:"longitude"`
	CoverImageURL string             `json:"cover_image_url" binding:"max=255"`
	Images        []string           `json:"images"`
	MenuImages    []string           `json:"menu_images"`
	Hours         []StoreHourRequest `json:"hours"`
	CategoryIDs   []int64            `json:"category_ids"`
}

// UpdateStoreRequest is the request payload for partially updating a store.
type UpdateStoreRequest struct {
	Name          *string             `json:"name" binding:"omitempty,max=255"`
	Description   *string             `json:"description" binding:"omitempty,max=2000"`
	Address       *string             `json:"address" binding:"omitempty,max=255"`
	City          *string             `json:"city" binding:"omitempty,max=100"`
	State         *string             `json:"state" binding:"omitempty,max=100"`
	ZipCode       *string             `json:"zip_code" binding:"omitempty,max=20"`
	Country       *string             `json:"country" binding:"omitempty,max=50"`
	Phone         *string             `json:"phone" binding:"omitempty,max=50"`
	Website       *string             `json:"website" binding:"omitempty,max=255"`
	Latitude      *float64            `json:"latitude"`
	Longitude     *float64            `json:"longitude"`
	CoverImageURL *string             `json:"cover_image_url" binding:"omitempty,max=255"`
	Images        *[]string           `json:"images"`
	MenuImages    *[]string           `json:"menu_images"`
	Hours         *[]StoreHourRequest `json:"hours"`
	CategoryIDs   *[]int64            `json:"category_ids"`
}

type StoreHourRequest struct {
	DayOfWeek int16  `json:"day_of_week"`
	OpenTime  string `json:"open_time" binding:"max=10"`
	CloseTime string `json:"close_time" binding:"max=10"`
	IsClosed  bool   `json:"is_closed"`
}

// StoreListQuery controls public store list filters.
type StoreListQuery struct {
	Category *string  // category name or id
	Lat      *float64 // latitude for proximity filter
	Lng      *float64 // longitude for proximity filter
	Rating   *float32 // min average rating
	RadiusKM *float64 // optional radius in KM for location filter
	Cursor   *int64   // id cursor for pagination (id < cursor)
	Limit    *int     // max rows
}

// StoreReviewListQuery controls public store reviews pagination.
type StoreReviewListQuery struct {
	Cursor *int64 // id cursor for pagination (id < cursor)
	Limit  *int   // max rows
}
