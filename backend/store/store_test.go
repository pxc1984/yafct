package store

import (
	"testing"

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
	testCardSetSessionFlow(s.T(), s.s)
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
