package dto

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
