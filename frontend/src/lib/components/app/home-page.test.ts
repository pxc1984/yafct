import { fireEvent, render, screen } from '@testing-library/svelte'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { CardImage } from '$lib/api/flashcards'

import HomePage from './home-page.svelte'

function parseCardData(input: string) {
  const blocks = input
    .trim()
    .split(/\n\s*\n+/)
    .map((block) => block.trim())
    .filter(Boolean)

  if (blocks.length === 0) {
    throw new Error('Добавь хотя бы одну карточку.')
  }

  return blocks.map((block, index) => {
    const lines = block
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)

    let question = ''
    let answer = ''
    let remarks = ''
    const questionImages: CardImage[] = []
    const answerImages: CardImage[] = []

    const readFieldValue = (line: string, prefix: string) => line.slice(prefix.length).trim()

    for (const line of lines) {
      if (line.startsWith('QUESTION::')) {
        question = readFieldValue(line, 'QUESTION::')
        continue
      }

      if (line.startsWith('ANSWER::')) {
        answer = readFieldValue(line, 'ANSWER::')
        continue
      }

      if (line.startsWith('REMARK::')) {
        remarks = readFieldValue(line, 'REMARK::')
        continue
      }

      if (line.startsWith('QUESTION_IMAGE::')) {
        questionImages.push({ id: readFieldValue(line, 'QUESTION_IMAGE::'), mimeType: 'image/png', dataBase64: 'aGVsbG8=' })
        continue
      }

      if (line.startsWith('ANSWER_IMAGE::')) {
        answerImages.push({ id: readFieldValue(line, 'ANSWER_IMAGE::'), mimeType: 'image/png', dataBase64: 'aGVsbG8=' })
        continue
      }

      throw new Error(`Карточка ${index + 1} содержит строку в неверном формате.`)
    }

    if (!question || !answer) {
      throw new Error(`У карточки ${index + 1} обязательны QUESTION и ANSWER.`)
    }

    return { question, answer, remarks, questionImages, answerImages }
  })
}

function buildSourceText(count: number) {
  return Array.from({ length: count }, (_, index) => {
    const number = index + 1
    return `QUESTION:: Question ${number}\nANSWER:: Answer ${number}\nREMARK:: Remark ${number}`
  }).join('\n\n')
}

const defaultProps = {
  promptText: 'prompt',
  setTitle: '',
  setDescription: '',
  setAuthor: '',
  parseCardData,
  resolveImageById: () => null,
  isCreating: false,
  createError: '',
  copyState: 'idle' as const,
  loadLinkError: '',
  onCopyPrompt: () => {},
  onLoadLink: () => {},
  onUploadImage: async () => ({ id: 'img-1', mimeType: 'image/png', dataBase64: 'aGVsbG8=' }),
  onCreateSet: () => {},
}

describe('HomePage cards list preview', () => {
  it('keeps the last question outside the visible list viewport for 50 cards', async () => {
    render(HomePage, {
      props: {
        sourceText: buildSourceText(50),
        ...defaultProps,
      },
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Показать список' }))

    const viewport = screen.getByTestId('cards-list-container')
    const lastCard = screen.getByTestId('preview-card-49')
    const lastQuestion = screen.getByDisplayValue('Question 50')

    Object.defineProperty(viewport, 'clientHeight', { configurable: true, value: 448 })
    Object.defineProperty(viewport, 'scrollHeight', { configurable: true, value: 5000 })

    viewport.getBoundingClientRect = () =>
      ({ top: 0, bottom: 448, left: 0, right: 320, width: 320, height: 448, x: 0, y: 0, toJSON() {} })

    lastCard.getBoundingClientRect = () =>
      ({ top: 4700, bottom: 4796, left: 0, right: 320, width: 320, height: 96, x: 0, y: 4700, toJSON() {} })

    expect(lastQuestion).toBeInTheDocument()
    expect(viewport.scrollHeight).toBeGreaterThan(viewport.clientHeight)
    expect(lastCard.getBoundingClientRect().top).toBeGreaterThan(viewport.getBoundingClientRect().bottom)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('adds a new card from list mode', async () => {
    render(HomePage, {
      props: {
        sourceText: buildSourceText(2),
        ...defaultProps,
      },
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Показать список' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Добавить карточку' }))

    expect(screen.getByTestId('preview-card-2')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Новый вопрос 3')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Новый ответ 3')).toBeInTheDocument()
  })

  it('parses generated cards back when remark is empty', () => {
    expect(() => parseCardData('QUESTION:: Новый вопрос 1\nANSWER:: Новый ответ 1\nREMARK:: ')).not.toThrow()
  })

  it('deletes a card only after second click within 4 seconds', async () => {
    vi.useFakeTimers()

    render(HomePage, {
      props: {
        sourceText: buildSourceText(2),
        ...defaultProps,
      },
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Показать список' }))

    const deleteButton = screen.getAllByRole('button', { name: 'Удалить карточку' })[0]
    await fireEvent.click(deleteButton)

    expect(screen.getByRole('button', { name: 'Подтвердить удаление карточки' })).toBeInTheDocument()
    expect(screen.getByTestId('preview-card-1')).toBeInTheDocument()

    await fireEvent.click(screen.getByRole('button', { name: 'Подтвердить удаление карточки' }))

    expect(screen.queryByTestId('preview-card-1')).not.toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Удалить карточку' })).toHaveLength(1)
  })

  it('cancels delete confirmation after 4 seconds', async () => {
    vi.useFakeTimers()

    render(HomePage, {
      props: {
        sourceText: buildSourceText(2),
        ...defaultProps,
      },
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Показать список' }))
    await fireEvent.click(screen.getAllByRole('button', { name: 'Удалить карточку' })[0])

    await vi.advanceTimersByTimeAsync(4000)

    expect(screen.queryByRole('button', { name: 'Подтвердить удаление карточки' })).not.toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Удалить карточку' })).toHaveLength(2)
  })
})
