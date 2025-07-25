package models

type log struct {
	ID        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int    `json:"user_id" gorm:"not null"`
	Action    string `json:"action" gorm:"not null"`
	Content   string `json:"content" gorm:"not null"`
	Timestamp string `json:"timestamp" gorm:"not null"`
	Hex       string `json:"hex" gorm:"type:text"`
}
