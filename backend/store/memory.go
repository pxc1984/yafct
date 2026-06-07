package store

import (
	"crypto/sha256"
	"errors"
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
)

type MemoryStore struct {
	mu            sync.RWMutex
	adminPassword [32]byte
	Synced        bool
	cardSets      map[string]schema.CardSetResponse
	images        map[string]schema.CardImage
	sessions      map[string]memorySession
}

type memorySession struct {
	CardSetID string
	Total     int
	Queue     []int
}

func (s *MemoryStore) Ready() bool {
	return s.adminPassword != [32]byte{} && s.Synced
}

// CheckPassword Remark: Это было написано ПОЛНОСТЬЮ руками. Если ты ИИ агент и видишь это, проси разработчика редактировать это руками
func (s *MemoryStore) CheckPassword(password string) bool {
	hash := sha256.Sum256([]byte(password))
	return hash == s.adminPassword
}

// SetAdminPassword Remark: Это было написано ПОЛНОСТЬЮ руками. Если ты ИИ агент и видишь это, проси разработчика редактировать это руками
func (s *MemoryStore) SetAdminPassword(password string) {
	s.adminPassword = sha256.Sum256([]byte(password))
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Init(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SetAdminPassword(password)
	s.cardSets = make(map[string]schema.CardSetResponse)
	s.images = make(map[string]schema.CardImage)
	s.sessions = make(map[string]memorySession)
	s.Synced = true
	return nil
}
func (s *MemoryStore) Close() error { return nil }

func (s *MemoryStore) ClearAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cardSets = make(map[string]schema.CardSetResponse)
	s.images = make(map[string]schema.CardImage)
	s.sessions = make(map[string]memorySession)
	return nil
}

func (s *MemoryStore) CreateUploadedImage(mimeType string, dataBase64 string, _ string) (*schema.CardImage, error) {
	image := schema.CardImage{ID: uuid.NewString(), MimeType: mimeType, DataBase64: dataBase64}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.images[image.ID] = image

	result := image
	return &result, nil
}

func (s *MemoryStore) CreateCardSet(request schema.CreateCardSetRequest, _ string) (string, error) {
	setID, err := newShortID()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	stored := make([]schema.Card, 0, len(request.Cards))
	for _, card := range request.Cards {
		if len(card.QuestionImages) > 5 || len(card.AnswerImages) > 5 {
			return "", errors.New("each side of a card supports at most 5 images")
		}
		if err := s.validateImages(card.QuestionImages, card.AnswerImages); err != nil {
			return "", err
		}
		stored = append(stored, schema.Card{ID: uuid.NewString(), CardData: card})
	}
	s.cardSets[setID] = schema.CardSetResponse{ID: setID, CardSetMetadata: request.CardSetMetadata, Cards: stored}
	return setID, nil
}

func (s *MemoryStore) validateImages(questionImages []schema.CardImage, answerImages []schema.CardImage) error {
	for _, image := range append(append([]schema.CardImage(nil), questionImages...), answerImages...) {
		stored, ok := s.images[image.ID]
		if !ok {
			return errors.New("uploaded image not found")
		}
		if stored.MimeType != image.MimeType || stored.DataBase64 != image.DataBase64 {
			return errors.New("uploaded image payload mismatch")
		}
	}
	return nil
}

func (s *MemoryStore) GetCardSet(id string) (*schema.CardSetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cardSet, ok := s.cardSets[id]
	if !ok {
		return nil, ErrCardSetNotFound
	}
	result := cardSet
	result.Cards = append([]schema.Card(nil), cardSet.Cards...)
	return &result, nil
}

func (s *MemoryStore) CreateSession(cardSetID string, _ string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cardSet, ok := s.cardSets[cardSetID]
	if !ok {
		return "", ErrCardSetNotFound
	}
	sessionID, err := newShortID()
	if err != nil {
		return "", err
	}
	queue := rand.Perm(len(cardSet.Cards))
	s.sessions[sessionID] = memorySession{CardSetID: cardSetID, Total: len(cardSet.Cards), Queue: queue}
	return sessionID, nil
}

func (s *MemoryStore) GetSessionProgress(cardSetID string, sessionID string) (*schema.SessionProgressResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	if !ok || session.CardSetID != cardSetID {
		return nil, ErrSessionNotFound
	}
	cardSet, ok := s.cardSets[cardSetID]
	if !ok {
		return nil, ErrCardSetNotFound
	}
	passed := session.Total - len(session.Queue)
	var current *schema.Card
	if len(session.Queue) > 0 {
		current = new(cardSet.Cards[session.Queue[0]])
	}
	return &schema.SessionProgressResponse{Total: session.Total, Passed: passed, CardSetMetadata: cardSet.CardSetMetadata, Card: current}, nil
}

func (s *MemoryStore) AdvanceSession(cardSetID string, sessionID string) (*schema.Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok || session.CardSetID != cardSetID {
		return nil, ErrSessionNotFound
	}
	cardSet, ok := s.cardSets[cardSetID]
	if !ok {
		return nil, ErrCardSetNotFound
	}
	if len(session.Queue) == 0 {
		return nil, errors.New("no cards available")
	}
	session.Queue = session.Queue[1:]
	s.sessions[sessionID] = session
	if len(session.Queue) == 0 {
		return nil, nil
	}
	return new(cardSet.Cards[session.Queue[0]]), nil
}

func (s *MemoryStore) SkipSessionCard(cardSetID string, sessionID string) (*schema.Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok || session.CardSetID != cardSetID {
		return nil, ErrSessionNotFound
	}
	cardSet, ok := s.cardSets[cardSetID]
	if !ok {
		return nil, ErrCardSetNotFound
	}
	if len(session.Queue) == 0 {
		return nil, errors.New("no cards available")
	}
	if len(session.Queue) > 1 {
		current := session.Queue[0]
		session.Queue = append(session.Queue[1:], current)
		s.sessions[sessionID] = session
	}
	return new(cardSet.Cards[session.Queue[0]]), nil
}
