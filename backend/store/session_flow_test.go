package store

import (
	"testing"

	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
	"github.com/pxc1984/flashcards-trainer/backend/store/interfaces"
	"github.com/stretchr/testify/assert"
)

func testCardSetSessionFlow(t *testing.T, s interfaces.StoreBase) {
	image, err := s.CreateUploadedImage("image/png", "aGVsbG8=", "127.0.0.1")
	assert.NoError(t, err)

	setID, err := s.CreateCardSet(schema.CreateCardSetRequest{
		CardSetMetadata: schema.CardSetMetadata{Title: "set title", Description: "set description", Author: "set author"},
		Cards: []schema.CardData{
			{Question: "q1", Answer: "a1", QuestionImages: []schema.CardImage{*image}},
			{Question: "q2", Answer: "a2"},
		},
	}, "127.0.0.1")
	assert.NoError(t, err)

	cardSet, err := s.GetCardSet(setID)
	assert.NoError(t, err)
	assert.Equal(t, "set title", cardSet.Title)
	assert.Len(t, cardSet.Cards, 2)
	assert.Len(t, cardSet.Cards[0].QuestionImages, 1)
	assert.Equal(t, image.ID, cardSet.Cards[0].QuestionImages[0].ID)

	sessionID, err := s.CreateSession(setID, "127.0.0.1")
	assert.NoError(t, err)

	progress, err := s.GetSessionProgress(setID, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, 2, progress.Total)
	assert.Equal(t, 0, progress.Passed)
	assert.Contains(t, []string{"q1", "q2"}, progress.Card.Question)

	skipCard, err := s.SkipSessionCard(setID, sessionID)
	assert.NoError(t, err)
	assert.Contains(t, []string{"q1", "q2"}, skipCard.Question)

	progress, err = s.GetSessionProgress(setID, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, 0, progress.Passed)

	advanceCard, err := s.AdvanceSession(setID, sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, advanceCard)
	assert.Contains(t, []string{"q1", "q2"}, advanceCard.Question)

	progress, err = s.GetSessionProgress(setID, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, 1, progress.Passed)
	assert.NotNil(t, progress.Card)
	assert.Contains(t, []string{"q1", "q2"}, progress.Card.Question)

	lastCard, err := s.AdvanceSession(setID, sessionID)
	assert.NoError(t, err)
	assert.Nil(t, lastCard)

	progress, err = s.GetSessionProgress(setID, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, 2, progress.Passed)
	assert.Nil(t, progress.Card)
}
