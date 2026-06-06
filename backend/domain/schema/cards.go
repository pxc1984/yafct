package schema

type CardData struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
	Remarks  string `json:"remarks"`
}

type Card struct {
	ID string `json:"id"`
	CardData
}

type CreateCardSetResponse struct {
	ID string `json:"id"`
}

type StartSessionResponse struct {
	SessionID string `json:"session_id"`
}

type SessionProgressResponse struct {
	Total  int   `json:"total"`
	Passed int   `json:"passed"`
	Card   *Card `json:"card"`
}
