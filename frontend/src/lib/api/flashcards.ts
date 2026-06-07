import { api } from "$lib/api/client";

export type CardData = {
  question: string;
  answer: string;
  remarks: string;
  questionImages: CardImage[];
  answerImages: CardImage[];
};

export type CardImage = {
  id: string;
  mimeType: string;
  dataBase64: string;
};

export type CardSetMetadata = {
  title: string;
  description: string;
  author: string;
};

export type CreateCardSetRequest = CardSetMetadata & {
  cards: CardData[];
};

export type Card = CardData & {
  id: string;
};

export type CardSet = CardSetMetadata & {
  id: string;
  cards: Card[];
};

export type SessionState = {
  total: number;
  passed: number;
  title: string;
  description: string;
  author: string;
  card?: Card | null;
};

export async function getCardSet(cardsetId: string) {
  const { data } = await api.get<CardSet>(`/api/v1/cards/${cardsetId}`);
  return data;
}

export async function createCardSet(payload: CreateCardSetRequest) {
  const { data } = await api.post<{ id: string }>("/api/v1/cards", payload);
  return data;
}

export async function uploadImage(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  const { data } = await api.post<{ image: CardImage }>(
    "/api/v1/images",
    formData,
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );

  return data.image;
}

export async function startSession(cardsetId: string) {
  const { data } = await api.post<{ session_id: string }>(
    `/api/v1/cards/${cardsetId}`,
  );
  return data;
}

export async function getSessionState(cardsetId: string, sessionId: string) {
  const { data } = await api.get<SessionState>(
    `/api/v1/cards/${cardsetId}/${sessionId}`,
  );
  return data;
}

export async function passCurrentCard(cardsetId: string, sessionId: string) {
  const { data } = await api.post<Card>(
    `/api/v1/cards/${cardsetId}/${sessionId}`,
  );
  return data;
}

export async function skipCurrentCard(cardsetId: string, sessionId: string) {
  await api.delete(`/api/v1/cards/${cardsetId}/${sessionId}`);
}
