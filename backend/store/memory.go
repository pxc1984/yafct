package store

import (
	"crypto/sha256"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
)

type MemoryStore struct {
	mu            sync.RWMutex
	adminPassword [32]byte
	Synced        bool
	cardSets      map[string][]schema.Card
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
	s.cardSets = make(map[string][]schema.Card)
	s.sessions = make(map[string]memorySession)
	s.Synced = true
	return nil
}
func (s *MemoryStore) Close() error { return nil }

func (s *MemoryStore) ClearAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cardSets = make(map[string][]schema.Card)
	s.sessions = make(map[string]memorySession)
	return nil
}

func (s *MemoryStore) CreateCardSet(cards []schema.CardData) (string, error) {
	setID, err := newShortID()
	if err != nil {
		return "", err
	}

	stored := make([]schema.Card, 0, len(cards))
	for _, card := range cards {
		stored = append(stored, schema.Card{ID: uuid.NewString(), CardData: card})
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cardSets[setID] = stored
	return setID, nil
}

func (s *MemoryStore) GetCardSet(id string) ([]schema.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cards, ok := s.cardSets[id]
	if !ok {
		return nil, ErrCardSetNotFound
	}
	result := make([]schema.Card, len(cards))
	copy(result, cards)
	return result, nil
}

func (s *MemoryStore) CreateSession(cardSetID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cards, ok := s.cardSets[cardSetID]
	if !ok {
		return "", ErrCardSetNotFound
	}
	sessionID, err := newShortID()
	if err != nil {
		return "", err
	}
	queue := make([]int, len(cards))
	for i := range cards {
		queue[i] = i
	}
	s.sessions[sessionID] = memorySession{CardSetID: cardSetID, Total: len(cards), Queue: queue}
	return sessionID, nil
}

func (s *MemoryStore) GetSessionProgress(cardSetID string, sessionID string) (*schema.SessionProgressResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	if !ok || session.CardSetID != cardSetID {
		return nil, ErrSessionNotFound
	}
	cards, ok := s.cardSets[cardSetID]
	if !ok {
		return nil, ErrCardSetNotFound
	}
	passed := session.Total - len(session.Queue)
	var current *schema.Card
	if len(session.Queue) > 0 {
		current = new(cards[session.Queue[0]])
	}
	return &schema.SessionProgressResponse{Total: session.Total, Passed: passed, Card: current}, nil
}

func (s *MemoryStore) AdvanceSession(cardSetID string, sessionID string) (*schema.Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok || session.CardSetID != cardSetID {
		return nil, ErrSessionNotFound
	}
	cards, ok := s.cardSets[cardSetID]
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
	return new(cards[session.Queue[0]]), nil
}

func (s *MemoryStore) SkipSessionCard(cardSetID string, sessionID string) (*schema.Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok || session.CardSetID != cardSetID {
		return nil, ErrSessionNotFound
	}
	cards, ok := s.cardSets[cardSetID]
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
	return new(cards[session.Queue[0]]), nil
}
