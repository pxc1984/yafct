package models

type CardSet struct {
	ID    string `gorm:"primaryKey;size:21"`
	Cards []Card `gorm:"constraint:OnDelete:CASCADE"`
}

type Card struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	CardSetID string `gorm:"index;size:21;not null"`
	Position  int    `gorm:"not null"`
	Question  string `gorm:"not null"`
	Answer    string `gorm:"not null"`
	Remarks   string
}

type CardSession struct {
	ID         string `gorm:"primaryKey;size:21"`
	CardSetID  string `gorm:"index;size:21;not null"`
	TotalCards int    `gorm:"not null"`
	Queue      string `gorm:"type:text;not null"`
}
