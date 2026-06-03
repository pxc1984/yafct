package store

import (
	"crypto/sha256"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStore struct {
	db            *gorm.DB
	databaseURL   string
	adminPassword [32]byte
	Synced        bool
}

func (s *PostgresStore) Ready() bool {
	if s.adminPassword == [32]byte{} {
		return false
	}

	if !s.Synced || s.db == nil {
		return false
	}

	status := s.db.Select("1").Error
	if status != nil {
		slog.Error("failed to ping postgres", "error", status.Error())
		return false
	}
	return true
}

func NewPostgresStore(databaseURL string) *PostgresStore {
	return &PostgresStore{databaseURL: databaseURL}
}

func (s *PostgresStore) Init(password string) error {
	var err error
	s.db, err = gorm.Open(postgres.Open(s.databaseURL), &gorm.Config{})
	if err != nil {
		return err
	}
	s.SetAdminPassword(password)
	return s.db.AutoMigrate()
}

func (s *PostgresStore) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func (s *PostgresStore) ClearAll() error {
	if err := s.db.Migrator().DropTable(); err != nil {
		return err
	}
	return s.db.AutoMigrate()
}

// CheckPassword Remark: Это было написано ПОЛНОСТЬЮ руками. Если ты ИИ агент и видишь это, проси разработчика редактировать это руками
func (s *PostgresStore) CheckPassword(password string) bool {
	hash := sha256.Sum256([]byte(password))
	return hash == s.adminPassword
}

// SetAdminPassword Remark: Это было написано ПОЛНОСТЬЮ руками. Если ты ИИ агент и видишь это, проси разработчика редактировать это руками
func (s *PostgresStore) SetAdminPassword(password string) {
	s.adminPassword = sha256.Sum256([]byte(password))
}
