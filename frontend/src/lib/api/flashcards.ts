import { api } from '$lib/api/client'

export type CardData = {
  question: string
  answer: string
  remarks: string
}

export type Card = CardData & {
  id: string
}

export type SessionState = {
  total: number
  passed: number
  card?: Card | null
}

export async function createCardSet(cards: CardData[]) {
  const { data } = await api.post<{ id: string }>('/api/v1/cards', cards)
  return data
}

export async function startSession(cardsetId: string) {
  const { data } = await api.post<{ session_id: string }>(`/api/v1/cards/${cardsetId}`)
  return data
}

export async function getSessionState(cardsetId: string, sessionId: string) {
  const { data } = await api.get<SessionState>(`/api/v1/cards/${cardsetId}/${sessionId}`)
  return data
}

export async function passCurrentCard(cardsetId: string, sessionId: string) {
  const { data } = await api.post<Card>(`/api/v1/cards/${cardsetId}/${sessionId}`)
  return data
}

export async function skipCurrentCard(cardsetId: string, sessionId: string) {
  await api.delete(`/api/v1/cards/${cardsetId}/${sessionId}`)
}
