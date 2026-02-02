package dto

// UserBrief is a minimal user representation for follow lists.
type UserBrief struct {
	UserID    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Intro     string `json:"intro"`
}

// FollowingUsersResponse returns users the current user follows.
type FollowingUsersResponse struct {
	Users []UserBrief `json:"users"`
	Total int         `json:"total"`
}

// FollowersResponse returns followers of the current user.
type FollowersResponse struct {
	Users []UserBrief `json:"users"`
	Total int         `json:"total"`
}
