package dto

import "time"

type FavoriteItem struct {
	ID         int64          `json:"id"`
	TargetType string         `json:"target_type"`
	TargetID   int64          `json:"target_id"`
	Post       *PostItem      `json:"post,omitempty"`
	Review     *ReviewItem    `json:"review,omitempty"`
	Merchant   *MerchantBrief `json:"merchant,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type FavoriteListResponse struct {
	Items  []FavoriteItem `json:"items"`
	Total  int            `json:"total"`
	Cursor *int64         `json:"cursor,omitempty"`
}
