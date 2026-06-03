const defaultApiUrl = 'http://localhost:8080'

export const API_URL = import.meta.env.API_URL?.trim() || defaultApiUrl
