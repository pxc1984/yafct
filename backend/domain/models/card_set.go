package models

import "github.com/lib/pq"

type CardSet struct {
	ID          string `gorm:"primaryKey;size:21"`
	Title       string `gorm:"size:120;not null;default:''"`
	Description string `gorm:"type:text;not null;default:''"`
	Author      string `gorm:"size:120;not null;default:''"`
	CreatedByIP string `gorm:"size:45;not null;default:''"`
	Cards       []Card `gorm:"constraint:OnDelete:CASCADE"`
}

type UploadedImage struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	MimeType    string `gorm:"size:255;not null"`
	DataBase64  string `gorm:"type:text;not null"`
	CreatedByIP string `gorm:"size:45;not null;default:''"`
}

type CardImage struct {
	ID         string `json:"id"`
	MimeType   string `json:"mimeType"`
	DataBase64 string `json:"dataBase64"`
}

type Card struct {
	ID             string `gorm:"primaryKey;type:uuid"`
	CardSetID      string `gorm:"index;size:21;not null"`
	Position       int    `gorm:"not null"`
	Question       string `gorm:"not null"`
	Answer         string `gorm:"not null"`
	Remarks        string
	QuestionImages []CardImage `gorm:"serializer:json;type:jsonb;not null;default:'[]'"`
	AnswerImages   []CardImage `gorm:"serializer:json;type:jsonb;not null;default:'[]'"`
}

type CardSession struct {
	ID          string        `gorm:"primaryKey;size:21"`
	CardSetID   string        `gorm:"index;size:21;not null"`
	CreatedByIP string        `gorm:"size:45;not null;default:''"`
	TotalCards  int           `gorm:"not null"`
	Queue       pq.Int64Array `gorm:"type:integer[]"`
	Current     int           `gorm:"default:-1"`
}
