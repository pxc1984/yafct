package interfaces

import (
	"github.com/google/uuid"
	"github.com/pxc1984/flashcards-trainer/backend/domain/models"
	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
)

type CacheStoreBase interface {
	Init() error
	Close() error
	ClearAll() error
	Ready() bool

	CreateAuthToken() (*models.AuthToken, error)
	CheckAuthToken(token uuid.UUID) (bool, error)
	DeleteAuthToken(token uuid.UUID) error
	Get(key string) ([]byte, bool, error)
	Set(key string, value []byte, ttlSeconds int64) error
	DeleteByPrefix(prefix string) error
	AllowRateLimit(key string, capacity int, refillPerSecond float64) (bool, float64, error)
}

type StoreBase interface {
	Init(password string) error
	Close() error
	ClearAll() error
	Ready() bool

	SetAdminPassword(password string)
	CheckPassword(password string) bool
	CreateCardSet(request schema.CreateCardSetRequest, createdByIP string) (string, error)
	GetCardSet(id string) (*schema.CardSetResponse, error)
	CreateSession(cardSetID string, createdByIP string) (string, error)
	GetSessionProgress(cardSetID string, sessionID string) (*schema.SessionProgressResponse, error)
	AdvanceSession(cardSetID string, sessionID string) (*schema.Card, error)
	SkipSessionCard(cardSetID string, sessionID string) (*schema.Card, error)
}
