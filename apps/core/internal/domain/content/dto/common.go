package dto

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
