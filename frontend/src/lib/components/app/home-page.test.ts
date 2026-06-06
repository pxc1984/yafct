import { fireEvent, render, screen } from '@testing-library/svelte'
import { describe, expect, it } from 'vitest'

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

    for (const line of lines) {
      if (line.startsWith('QUESTION:: ')) {
        question = line.slice('QUESTION:: '.length).trim()
        continue
      }

      if (line.startsWith('ANSWER:: ')) {
        answer = line.slice('ANSWER:: '.length).trim()
        continue
      }

      if (line.startsWith('REMARK:: ')) {
        remarks = line.slice('REMARK:: '.length).trim()
        continue
      }

      throw new Error(`Карточка ${index + 1} содержит строку в неверном формате.`)
    }

    if (!question || !answer) {
      throw new Error(`У карточки ${index + 1} обязательны QUESTION и ANSWER.`)
    }

    return { question, answer, remarks }
  })
}

function buildSourceText(count: number) {
  return Array.from({ length: count }, (_, index) => {
    const number = index + 1
    return `QUESTION:: Question ${number}\nANSWER:: Answer ${number}\nREMARK:: Remark ${number}`
  }).join('\n\n')
}

describe('HomePage cards list preview', () => {
  it('keeps the last question outside the visible list viewport for 50 cards', async () => {
    render(HomePage, {
      props: {
        promptText: 'prompt',
        sourceText: buildSourceText(50),
        setTitle: '',
        setDescription: '',
        setAuthor: '',
        parseCardData,
        isCreating: false,
        createError: '',
        copyState: 'idle',
        onCopyPrompt: () => {},
        onCreateSet: () => {},
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
})
