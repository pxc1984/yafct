package store

import (
	"os"
	"testing"

	"github.com/pxc1984/flashcards-trainer/backend/store/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PostgresStoreTestSuite struct {
	suite.Suite
	s interfaces.StoreBase
}

func (s *PostgresStoreTestSuite) SetupTest() {
	url := os.Getenv("TEST_POSTGRES_URL")
	if url == "" {
		s.T().Skip("TEST_POSTGRES_URL not set; skipping postgres store tests")
	}
	store := NewPostgresStore(url)
	assert.NoError(s.T(), store.Init("admin"))
	assert.NoError(s.T(), store.ClearAll())
	s.s = store
}

func (s *PostgresStoreTestSuite) TearDownTest() {
	if s.s != nil {
		assert.NoError(s.T(), s.s.ClearAll())
		assert.NoError(s.T(), s.s.Close())
	}
}

func (s *PostgresStoreTestSuite) TestCardSetSessionFlow() {
	testCardSetSessionFlow(s.T(), s.s)
}

func TestPostgresStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresStoreTestSuite))
}
