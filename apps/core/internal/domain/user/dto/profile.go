package dto

type ProfileResponse struct {
	UserID    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Intro     string `json:"intro"`
	Location  string `json:"location"`
}

type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Intro     *string `json:"intro,omitempty"`
	Location  *string `json:"location,omitempty"`
}
