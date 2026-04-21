package dto

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

type VoucherResponse struct {
	ID             int64      `json:"id"`
	Code           string     `json:"code"`
	CouponID       int64      `json:"coupon_id"`
	UserID         int64      `json:"user_id"`
	PackageID      *int64     `json:"package_id"`
	OrderID        *int64     `json:"order_id"`
	MerchantID     *int64     `json:"merchant_id"`
	Status         string     `json:"status"`
	ExpiryDate     time.Time  `json:"expiry_date"`
	ValidFrom      *time.Time `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	RedeemedAt     *time.Time `json:"redeemed_at"`
	RedeemedBy     *int64     `json:"redeemed_by"`
	RedemptionNote string     `json:"redemption_note"`
	CouponTitle    string     `json:"coupon_title,omitempty"`
	MerchantName   string     `json:"merchant_name,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ScanURL        string     `json:"scan_url"`
}

func FromModel(v model.Voucher, frontendURL string) (VoucherResponse, error) {
	scanURL, err := BuildScanURL(frontendURL, v.ScanToken)
	if err != nil {
		return VoucherResponse{}, err
	}

	return VoucherResponse{
		ID:             v.ID,
		Code:           v.Code,
		CouponID:       v.CouponID,
		UserID:         v.UserID,
		PackageID:      v.PackageID,
		OrderID:        v.OrderID,
		MerchantID:     v.MerchantID,
		Status:         v.Status,
		ExpiryDate:     v.ExpiryDate,
		ValidFrom:      v.ValidFrom,
		ValidUntil:     v.ValidUntil,
		RedeemedAt:     v.RedeemedAt,
		RedeemedBy:     v.RedeemedBy,
		RedemptionNote: v.RedemptionNote,
		CouponTitle:    v.CouponTitle,
		MerchantName:   v.MerchantName,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
		ScanURL:        scanURL,
	}, nil
}

func BuildScanURL(frontendURL, scanToken string) (string, error) {
	trimmedBase := strings.TrimRight(frontendURL, "/")
	if trimmedBase == "" {
		return "", fmt.Errorf("frontend url is empty")
	}
	if strings.TrimSpace(scanToken) == "" {
		return "", fmt.Errorf("voucher scan token is empty")
	}
	return trimmedBase + "/merchant/vouchers/scan?t=" + url.QueryEscape(scanToken), nil
}
