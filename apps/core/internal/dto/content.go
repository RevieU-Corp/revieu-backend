package dto

import "time"

type UserBrief struct {
	UserID    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Intro     string `json:"intro"`
}

type FollowingUsersResponse struct {
	Users []UserBrief `json:"users"`
	Total int         `json:"total"`
}

type FollowersResponse struct {
	Users []UserBrief `json:"users"`
	Total int         `json:"total"`
}

type MerchantBrief struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

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

type ReviewItem struct {
	ID            int64         `json:"id"`
	Rating        float32       `json:"rating"`
	RatingEnv     *float32      `json:"rating_env,omitempty"`
	RatingService *float32      `json:"rating_service,omitempty"`
	RatingValue   *float32      `json:"rating_value,omitempty"`
	Content       string        `json:"content"`
	Images        []string      `json:"images"`
	AvgCost       *int          `json:"avg_cost,omitempty"`
	LikeCount     int           `json:"like_count"`
	IsLiked       bool          `json:"is_liked"`
	Merchant      MerchantBrief `json:"merchant"`
	Tags          []string      `json:"tags"`
	CreatedAt     time.Time     `json:"created_at"`
}

type ReviewListResponse struct {
	Reviews []ReviewItem `json:"reviews"`
	Total   int          `json:"total"`
	Cursor  *int64       `json:"cursor,omitempty"`
}

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
