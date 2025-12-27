package model

import "time"

type LoginAttempt struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Email       string    `gorm:"index;size:255;not null" json:"email"`
	IPAddress   string    `gorm:"size:45" json:"ip_address"`
	AttemptTime time.Time `gorm:"not null;index" json:"attempt_time"`
	Success     bool      `gorm:"default:false" json:"success"`
	CreatedAt   time.Time `json:"created_at"`
}


func (LoginAttempt) TableName() string {
	return "login_attempts"
}