/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly API_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.txt?raw' {
  const content: string
  export default content
}
