package dto

type PublicProfileResponse struct {
	UserID         int64  `json:"user_id"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	Intro          string `json:"intro"`
	Location       string `json:"location"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	PostCount      int    `json:"post_count"`
	ReviewCount    int    `json:"review_count"`
	LikeCount      int    `json:"like_count"`
	IsFollowing    bool   `json:"is_following"`
}

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

type SecurityOverviewResponse struct {
	HasPassword    bool     `json:"has_password"`
	LinkedAccounts []string `json:"linked_accounts"`
	Email          string   `json:"email"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type PrivacySettings struct {
	IsPublic bool `json:"is_public"`
}

type NotificationSettings struct {
	PushEnabled  bool `json:"push_enabled"`
	EmailEnabled bool `json:"email_enabled"`
}

type AddressItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Province   string `json:"province"`
	City       string `json:"city"`
	District   string `json:"district"`
	Address    string `json:"address"`
	PostalCode string `json:"postal_code"`
	IsDefault  bool   `json:"is_default"`
}

type AddressListResponse struct {
	Addresses []AddressItem `json:"addresses"`
}

type CreateAddressRequest struct {
	Name       string `json:"name" binding:"required,max=50"`
	Phone      string `json:"phone" binding:"required,max=20"`
	Province   string `json:"province" binding:"max=50"`
	City       string `json:"city" binding:"max=50"`
	District   string `json:"district" binding:"max=50"`
	Address    string `json:"address" binding:"required,max=255"`
	PostalCode string `json:"postal_code" binding:"max=20"`
	IsDefault  bool   `json:"is_default"`
}

type UpdateAddressRequest struct {
	Name       *string `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Province   *string `json:"province,omitempty"`
	City       *string `json:"city,omitempty"`
	District   *string `json:"district,omitempty"`
	Address    *string `json:"address,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	IsDefault  *bool   `json:"is_default,omitempty"`
}
