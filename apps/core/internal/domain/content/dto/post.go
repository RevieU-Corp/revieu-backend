package dto

import "time"

type PostItem struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Images    []string       `json:"images"`
	LikeCount int            `json:"like_count"`
	ViewCount int            `json:"view_count"`
	IsLiked   bool           `json:"is_liked"`
	Merchant  *MerchantBrief `json:"merchant,omitempty"`
	Tags      []string       `json:"tags"`
	CreatedAt time.Time      `json:"created_at"`
}

type PostListResponse struct {
	Posts  []PostItem `json:"posts"`
	Total  int        `json:"total"`
	Cursor *int64     `json:"cursor,omitempty"`
}
