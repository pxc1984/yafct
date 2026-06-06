package cards

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
	"github.com/pxc1984/flashcards-trainer/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardsAPIFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeObj := store.NewMemoryStore()
	require.NoError(t, storeObj.Init("admin"))

	router := gin.New()
	apiGroup := router.Group("/api/v1")
	RegisterRoutes(apiGroup, storeObj)

	createBody := mustJSON(t, []schema.CardData{
		{Question: "q1", Answer: "a1", Remarks: "r1"},
		{Question: "q2", Answer: "a2", Remarks: "r2"},
	})
	createResp := performRequest(router, http.MethodPost, "/api/v1/cards", createBody)
	require.Equal(t, http.StatusCreated, createResp.Code)

	var created schema.CreateCardSetResponse
	require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
	assert.Len(t, created.ID, 8)

	getSetResp := performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID, nil)
	require.Equal(t, http.StatusOK, getSetResp.Code)

	var cardsResp []schema.Card
	require.NoError(t, json.Unmarshal(getSetResp.Body.Bytes(), &cardsResp))
	require.Len(t, cardsResp, 2)
	assert.Equal(t, "q1", cardsResp[0].Question)
	assert.Equal(t, "q2", cardsResp[1].Question)

	startResp := performRequest(router, http.MethodPost, "/api/v1/cards/"+created.ID, nil)
	require.Equal(t, http.StatusOK, startResp.Code)

	var session schema.StartSessionResponse
	require.NoError(t, json.Unmarshal(startResp.Body.Bytes(), &session))
	assert.Len(t, session.SessionID, 8)

	progressResp := performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)

	var progress schema.SessionProgressResponse
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 2, progress.Total)
	assert.Equal(t, 0, progress.Passed)
	require.NotNil(t, progress.Card)
	assert.Equal(t, "q1", progress.Card.Question)

	advanceResp := performRequest(router, http.MethodPost, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusSeeOther, advanceResp.Code)
	assert.Equal(t, "/api/v1/cards/"+created.ID+"/"+session.SessionID, advanceResp.Header().Get("Location"))

	progressResp = performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 1, progress.Passed)
	require.NotNil(t, progress.Card)
	assert.Equal(t, "q2", progress.Card.Question)

	skipResp := performRequest(router, http.MethodDelete, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusSeeOther, skipResp.Code)
	assert.Equal(t, "/api/v1/cards/"+created.ID+"/"+session.SessionID, skipResp.Header().Get("Location"))

	progressResp = performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 1, progress.Passed)
	require.NotNil(t, progress.Card)
	assert.Equal(t, "q2", progress.Card.Question)

	advanceResp = performRequest(router, http.MethodPost, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusSeeOther, advanceResp.Code)

	progressResp = performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 2, progress.Passed)
	assert.Nil(t, progress.Card)
}

func TestCardsAPIErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeObj := store.NewMemoryStore()
	require.NoError(t, storeObj.Init("admin"))

	router := gin.New()
	apiGroup := router.Group("/api/v1")
	RegisterRoutes(apiGroup, storeObj)

	resp := performRequest(router, http.MethodPost, "/api/v1/cards", mustJSON(t, []schema.CardData{}))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	resp = performRequest(router, http.MethodGet, "/api/v1/cards/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	resp = performRequest(router, http.MethodPost, "/api/v1/cards/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	resp = performRequest(router, http.MethodGet, "/api/v1/cards/missing/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	resp = performRequest(router, http.MethodDelete, "/api/v1/cards/missing/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func performRequest(router http.Handler, method string, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.5:1234"
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}
