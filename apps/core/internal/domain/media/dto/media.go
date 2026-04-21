package dto

import "time"

type FileRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type PresignedURLRequest struct {
	Files []FileRequest `json:"files"`
}

type UploadInfo struct {
	ID        string    `json:"id"`
	UploadURL string    `json:"upload_url"`
	FileURL   string    `json:"file_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PresignedURLResponse struct {
	Uploads []UploadInfo `json:"uploads"`
}
