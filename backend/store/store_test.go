package store

import (
	"testing"

	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
	"github.com/pxc1984/flashcards-trainer/backend/store/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MemoryStoreTestSuite struct {
	suite.Suite
	s interfaces.StoreBase
}

func (s *MemoryStoreTestSuite) SetupTest() {
	s.s = NewMemoryStore()
	assert.Nil(s.T(), s.s.Init("admin"))
}

func (s *MemoryStoreTestSuite) TearDownTest() {
	assert.Nil(s.T(), s.s.Close())
}

func (s *MemoryStoreTestSuite) TestAdminPassword() {
	assert.True(s.T(), s.s.CheckPassword("admin"))
}

func (s *MemoryStoreTestSuite) TestCardSetSessionFlow() {
	image, err := s.s.CreateUploadedImage("image/png", "aGVsbG8=", "127.0.0.1")
	assert.NoError(s.T(), err)

	setID, err := s.s.CreateCardSet(schema.CreateCardSetRequest{
		CardSetMetadata: schema.CardSetMetadata{Title: "set title", Description: "set description", Author: "set author"},
		Cards: []schema.CardData{
			{Question: "q1", Answer: "a1", QuestionImages: []schema.CardImage{*image}},
			{Question: "q2", Answer: "a2"},
		},
	}, "127.0.0.1")
	assert.NoError(s.T(), err)

	cardSet, err := s.s.GetCardSet(setID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "set title", cardSet.Title)
	assert.Len(s.T(), cardSet.Cards, 2)
	assert.Len(s.T(), cardSet.Cards[0].QuestionImages, 1)
	assert.Equal(s.T(), image.ID, cardSet.Cards[0].QuestionImages[0].ID)

	sessionID, err := s.s.CreateSession(setID, "127.0.0.1")
	assert.NoError(s.T(), err)

	progress, err := s.s.GetSessionProgress(setID, sessionID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, progress.Total)
	assert.Equal(s.T(), 0, progress.Passed)
	assert.Equal(s.T(), "q1", progress.Card.Question)

	skipped, err := s.s.SkipSessionCard(setID, sessionID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "q2", skipped.Question)

	next, err := s.s.AdvanceSession(setID, sessionID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "q1", next.Question)
}

func TestMemoryStoreTestSuite(t *testing.T) {
	suite.Run(t, new(MemoryStoreTestSuite))
}

// example
//func (s *MemoryStoreTestSuite) TestCreateAndGetCourse() {
//	course := interfaces.CourseData{
//		ID:                  uuid.New(),
//		Title:               "Test Course",
//		Description:         new("A test course"),
//		HandbookLink:        new("http://example.com"),
//		AllowedCohorts:      []int{2024, 2025},
//		AvailableSemesters:  []int{1, 2},
//		RecommendedSemester: new(1),
//		Workload:            5.0,
//		CsatMetric:          new(4.5),
//	}
//
//	created, err := s.s.CreateCourse(course)
//	assert.NoError(s.T(), err)
//	assert.Equal(s.T(), course.ID, created.ID)
//
//	retrieved, err := s.s.GetCourseByID(course.ID)
//	assert.NoError(s.T(), err)
//	assert.NotNil(s.T(), retrieved)
//	assert.Equal(s.T(), "Test Course", retrieved.Title)
//	assert.Equal(s.T(), 5.0, retrieved.Workload)
//}
