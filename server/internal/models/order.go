package models

import "time"

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type (
	Order struct {
		ID         int       `json:"-"`
		Number     string    `json:"number"`
		UserID     int       `json:"-"`
		Status     string    `json:"status"`
		Accrual    *float64  `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded_at"`
	}
)
