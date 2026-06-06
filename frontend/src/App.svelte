<script lang="ts">
  import Check from '@lucide/svelte/icons/check'
  import ChevronRight from '@lucide/svelte/icons/chevron-right'
  import Copy from '@lucide/svelte/icons/copy'
  import GripHorizontal from '@lucide/svelte/icons/grip-horizontal'
  import Keyboard from '@lucide/svelte/icons/keyboard'
  import Play from '@lucide/svelte/icons/play'
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw'

  import promptTemplate from '../prompt.txt?raw'
  import {
    createCardSet,
    getSessionState,
    passCurrentCard,
    skipCurrentCard,
    startSession,
    type CardData,
    type SessionState,
  } from '$lib/api/flashcards'
  import RichMathText from '$lib/components/rich-math-text.svelte'
  import { Badge } from '$lib/components/ui/badge'
  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'

  type SessionRecord = {
    id: string
    updatedAt: string
  }

  type Route =
    | { name: 'home' }
    | { name: 'cardset'; cardsetId: string }
    | { name: 'session'; cardsetId: string; sessionId: string }

  const STORAGE_KEY = 'flashcards-trainer-sessions'
  const promptText = promptTemplate.trim()

  let route = $state<Route>(parseRoute(window.location.pathname))
  let sourceText = $state('')
  let isCreating = $state(false)
  let createError = $state('')
  let copyState = $state<'idle' | 'done'>('idle')

  let sessions = $state<SessionRecord[]>([])
  let sessionsLoading = $state(false)
  let sessionListError = $state('')

  let trainingState = $state<SessionState | null>(null)
  let trainingLoading = $state(false)
  let trainingError = $state('')
  let isAnswerVisible = $state(false)
  let dragOffset = $state(0)
  let isDragging = $state(false)
  let dragStartX = 0
  let isMobile = $state(false)

  const progressValue = $derived.by(() => {
    if (!trainingState || trainingState.total === 0) {
      return 0
    }

    return Math.min(100, Math.round((trainingState.passed / trainingState.total) * 100))
  })

  const hasActiveCard = $derived(Boolean(trainingState?.card))
  const activeCardsetId = $derived(route.name === 'home' ? '' : route.cardsetId)

  function parseRoute(pathname: string): Route {
    const parts = pathname.split('/').filter(Boolean)

    if (parts.length === 1) {
      return { name: 'cardset', cardsetId: parts[0] }
    }

    if (parts.length >= 2) {
      return { name: 'session', cardsetId: parts[0], sessionId: parts[1] }
    }

    return { name: 'home' }
  }

  function navigate(path: string) {
    window.history.pushState({}, '', path)
    route = parseRoute(path)
  }

  function readStoredSessions(cardsetId: string) {
    if (typeof localStorage === 'undefined') {
      return []
    }

    const raw = localStorage.getItem(STORAGE_KEY)

    if (!raw) {
      return []
    }

    try {
      const parsed = JSON.parse(raw) as Record<string, SessionRecord[]>
      return Array.isArray(parsed[cardsetId]) ? parsed[cardsetId] : []
    } catch {
      return []
    }
  }

  function storeSession(cardsetId: string, sessionId: string) {
    if (typeof localStorage === 'undefined') {
      return
    }

    const raw = localStorage.getItem(STORAGE_KEY)
    const parsed = raw ? (JSON.parse(raw) as Record<string, SessionRecord[]>) : {}
    const current = Array.isArray(parsed[cardsetId]) ? parsed[cardsetId] : []
    const next = [
      { id: sessionId, updatedAt: new Date().toISOString() },
      ...current.filter((session) => session.id !== sessionId),
    ]

    parsed[cardsetId] = next
    localStorage.setItem(STORAGE_KEY, JSON.stringify(parsed))
  }

  function syncSessions(cardsetId: string) {
    sessions = readStoredSessions(cardsetId)
  }

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

      return { question, answer, remarks } satisfies CardData
    })
  }

  async function copyPrompt() {
    await navigator.clipboard.writeText(promptText)
    copyState = 'done'

    window.setTimeout(() => {
      copyState = 'idle'
    }, 1500)
  }

  async function createSet() {
    createError = ''

    let cards: CardData[]

    try {
      cards = parseCardData(sourceText)
    } catch (error) {
      createError = error instanceof Error ? error.message : 'Не удалось разобрать карточки.'
      return
    }

    isCreating = true

    try {
      const { id } = await createCardSet(cards)
      navigate(`/${id}`)
    } catch (error) {
      createError = error instanceof Error ? error.message : 'Не удалось создать набор.'
    } finally {
      isCreating = false
    }
  }

  async function createSession(cardsetId: string) {
    sessionsLoading = true
    sessionListError = ''

    try {
      const { session_id } = await startSession(cardsetId)
      storeSession(cardsetId, session_id)
      syncSessions(cardsetId)
      navigate(`/${cardsetId}/${session_id}`)
    } catch (error) {
      sessionListError = error instanceof Error ? error.message : 'Не удалось создать сессию.'
    } finally {
      sessionsLoading = false
    }
  }

  async function loadTraining(cardsetId: string, sessionId: string) {
    trainingLoading = true
    trainingError = ''

    try {
      trainingState = await getSessionState(cardsetId, sessionId)
      isAnswerVisible = false
      dragOffset = 0
      storeSession(cardsetId, sessionId)
    } catch (error) {
      trainingError = error instanceof Error ? error.message : 'Не удалось загрузить карточку.'
    } finally {
      trainingLoading = false
    }
  }

  async function markKnown() {
    if (route.name !== 'session' || !trainingState?.card) {
      return
    }

    trainingLoading = true
    trainingError = ''

    try {
      await passCurrentCard(route.cardsetId, route.sessionId)
      await loadTraining(route.cardsetId, route.sessionId)
    } catch (error) {
      trainingError = error instanceof Error ? error.message : 'Не удалось перейти к следующей карточке.'
      trainingLoading = false
    }
  }

  async function markUnknown() {
    if (route.name !== 'session' || !trainingState?.card) {
      return
    }

    trainingLoading = true
    trainingError = ''

    try {
      await skipCurrentCard(route.cardsetId, route.sessionId)
      await loadTraining(route.cardsetId, route.sessionId)
    } catch (error) {
      trainingError = error instanceof Error ? error.message : 'Не удалось отложить карточку.'
      trainingLoading = false
    }
  }

  function handlePointerDown(event: PointerEvent) {
    if (!hasActiveCard || trainingLoading) {
      return
    }

    isDragging = true
    dragStartX = event.clientX
  }

  function handlePointerMove(event: PointerEvent) {
    if (!isDragging) {
      return
    }

    dragOffset = event.clientX - dragStartX
  }

  async function handlePointerUp() {
    if (!isDragging) {
      return
    }

    isDragging = false

    if (dragOffset >= 120) {
      dragOffset = 0
      await markKnown()
      return
    }

    if (dragOffset <= -120) {
      dragOffset = 0
      await markUnknown()
      return
    }

    dragOffset = 0
  }

  function toggleAnswer() {
    if (!trainingState?.card) {
      return
    }

    isAnswerVisible = !isAnswerVisible
  }

  function formatDate(value: string) {
    return new Date(value).toLocaleString('ru-RU', {
      dateStyle: 'short',
      timeStyle: 'short',
    })
  }

  $effect(() => {
    const onPopState = () => {
      route = parseRoute(window.location.pathname)
    }

    const media = window.matchMedia('(max-width: 767px)')
    const syncMobile = () => {
      isMobile = media.matches
    }

    syncMobile()
    window.addEventListener('popstate', onPopState)
    media.addEventListener('change', syncMobile)

    return () => {
      window.removeEventListener('popstate', onPopState)
      media.removeEventListener('change', syncMobile)
    }
  })

  $effect(() => {
    if (route.name === 'cardset') {
      syncSessions(route.cardsetId)
    }

    if (route.name === 'session') {
      void loadTraining(route.cardsetId, route.sessionId)
    }
  })

  $effect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (route.name !== 'session' || trainingLoading || !trainingState?.card) {
        return
      }

      if (event.code === 'Space') {
        event.preventDefault()
        toggleAnswer()
      }

      if (event.key.toLowerCase() === 'y') {
        event.preventDefault()
        void markKnown()
      }

      if (event.key.toLowerCase() === 'n') {
        event.preventDefault()
        void markUnknown()
      }
    }

    window.addEventListener('keydown', onKeyDown)

    return () => {
      window.removeEventListener('keydown', onKeyDown)
    }
  })
</script>

<main class="min-h-screen bg-background text-foreground">
  <div class="mx-auto flex min-h-screen w-full max-w-5xl flex-col px-4 py-6 sm:px-6 sm:py-8">
    {#if route.name === 'home'}
      <section class="mx-auto flex w-full flex-1 items-center">
        <Card.Root class="border-border/70 bg-card/85 shadow-sm backdrop-blur">
          <Card.Header class="gap-4">
            <Badge variant="secondary" class="w-fit">LLM prompt</Badge>
            <div class="space-y-2">
              <Card.Title class="text-3xl sm:text-4xl">Тренировка по карточкам</Card.Title>
              <Card.Description>
                Скопируй системный промпт, сгенерируй карточки в текстовом формате и вставь
                результат ниже.
              </Card.Description>
            </div>
          </Card.Header>
          <Card.Content class="space-y-4">
            <div class="relative rounded-2xl border bg-background/70 p-4 pr-16">
              <Button
                variant="outline"
                size="icon-sm"
                class="absolute top-3 right-3"
                onclick={copyPrompt}
                aria-label="Скопировать промпт"
              >
                <Copy class="size-4" />
              </Button>
              <pre class="max-h-[4.75rem] overflow-auto whitespace-pre-wrap text-sm text-muted-foreground">{promptText}</pre>
            </div>

            <label class="space-y-2">
              <span class="text-sm font-medium">Текст с карточками</span>
              <textarea
                bind:value={sourceText}
                class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 min-h-72 w-full rounded-2xl border bg-transparent px-4 py-3 text-sm outline-none focus-visible:ring-3"
                placeholder={`QUESTION:: Что такое closure?
ANSWER:: Функция вместе с лексическим окружением.
REMARK:: Удобно для инкапсуляции состояния.

QUESTION:: Что возвращает выражение $2^3$?
ANSWER:: $8$
REMARK:: `}
              ></textarea>
            </label>

            {#if createError}
              <p class="text-sm text-destructive">{createError}</p>
            {/if}
          </Card.Content>
          <Card.Footer class="justify-between gap-3 max-sm:flex-col max-sm:items-stretch">
            <p class="text-sm text-muted-foreground">
              {copyState === 'done' ? 'Промпт скопирован.' : ''}
            </p>
            <Button size="lg" onclick={createSet} disabled={isCreating || !sourceText.trim()}>
              {isCreating ? 'Создание...' : 'Создать набор'}
            </Button>
          </Card.Footer>
        </Card.Root>
      </section>
    {/if}

    {#if route.name === 'cardset'}
      <section class="mx-auto flex w-full max-w-4xl flex-1 items-center">
        <div class="grid w-full gap-6 lg:grid-cols-[0.95fr_1.05fr]">
          <Card.Root>
            <Card.Header>
              <Badge variant="outline" class="w-fit">Набор</Badge>
              <Card.Title class="break-all text-2xl">{route.cardsetId}</Card.Title>
              <Card.Description>
                Запусти новую тренировочную сессию или вернись к одной из уже начатых.
              </Card.Description>
            </Card.Header>
            <Card.Content>
              <Button size="lg" class="w-full gap-2" onclick={() => createSession(activeCardsetId)} disabled={sessionsLoading}>
                <Play class="size-4" />
                {sessionsLoading ? 'Создание...' : 'Начать новую сессию'}
              </Button>
              {#if sessionListError}
                <p class="mt-3 text-sm text-destructive">{sessionListError}</p>
              {/if}
            </Card.Content>
          </Card.Root>

          <Card.Root>
            <Card.Header>
              <Card.Title>Продолжить</Card.Title>
            </Card.Header>
            <Card.Content class="space-y-3">
              {#if sessions.length === 0}
                <p class="text-sm text-muted-foreground">Пока нет начатых сессий для этого набора.</p>
              {/if}

              {#each sessions as session}
                <button
                  class="flex w-full items-center justify-between rounded-2xl border bg-background/60 px-4 py-3 text-left transition hover:bg-muted/60"
                  onclick={() => navigate(`/${activeCardsetId}/${session.id}`)}
                >
                  <div>
                    <p class="font-medium">Сессия {session.id}</p>
                    <p class="text-sm text-muted-foreground">Обновлена {formatDate(session.updatedAt)}</p>
                  </div>
                  <ChevronRight class="size-4 text-muted-foreground" />
                </button>
              {/each}
            </Card.Content>
          </Card.Root>
        </div>
      </section>
    {/if}

    {#if route.name === 'session'}
      <section class="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-4 pb-28 sm:pb-8">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="text-sm text-muted-foreground">Набор {route.cardsetId}</p>
            <h1 class="text-2xl font-semibold">Тренировка</h1>
          </div>
          <Button variant="outline" onclick={() => navigate(`/${activeCardsetId}`)}>К сессиям</Button>
        </div>

        <Card.Root>
          <Card.Content class="space-y-3 pt-6">
            <div class="flex items-center justify-between text-sm text-muted-foreground">
              <span>Прогресс</span>
              <span>{trainingState?.passed ?? 0} / {trainingState?.total ?? 0}</span>
            </div>
            <div class="h-3 overflow-hidden rounded-full bg-muted">
              <div class="h-full rounded-full bg-primary transition-all" style={`width: ${progressValue}%`}></div>
            </div>
          </Card.Content>
        </Card.Root>

        <div class="grid gap-4 md:grid-cols-[1fr_220px]">
          <Card.Root
            class="overflow-hidden"
            onpointerdown={handlePointerDown}
            onpointermove={handlePointerMove}
            onpointerup={handlePointerUp}
            onpointercancel={handlePointerUp}
          >
            <div
              class={`h-full transition-transform ${isDragging ? '' : 'duration-200'}`}
              style={`transform: translateX(${dragOffset}px) rotate(${dragOffset / 30}deg);`}
            >
              <Card.Header>
                <div class="flex items-center justify-between gap-3">
                  <Badge variant="secondary">Вопрос</Badge>
                  <Badge variant={dragOffset > 50 ? 'secondary' : dragOffset < -50 ? 'destructive' : 'outline'}>
                    {dragOffset > 50 ? 'Знаю' : dragOffset < -50 ? 'Пока не знаю' : 'Смахни карточку'}
                  </Badge>
                </div>
              </Card.Header>
              <Card.Content class="space-y-6">
                {#if trainingLoading && !trainingState}
                  <p class="text-muted-foreground">Загрузка...</p>
                {:else if trainingError}
                  <p class="text-destructive">{trainingError}</p>
                {:else if trainingState?.card}
                  <div class="space-y-3">
                    <RichMathText text={trainingState.card.question} class="text-2xl leading-tight font-semibold sm:text-3xl" />
                    {#if trainingState.card.remarks}
                      <RichMathText text={trainingState.card.remarks} class="text-sm text-muted-foreground" />
                    {/if}
                  </div>

                  <div class="rounded-2xl border bg-background/60 p-4">
                    <p class="mb-2 text-sm font-medium text-muted-foreground">Ответ</p>
                    {#if isAnswerVisible}
                      <RichMathText text={trainingState.card.answer} class="text-lg leading-relaxed" />
                    {:else}
                      <p class="text-lg text-muted-foreground">Нажми показать ответ или пробел.</p>
                    {/if}
                  </div>
                {:else}
                  <div class="space-y-3 py-8 text-center">
                    <p class="text-2xl font-semibold">Сессия завершена</p>
                    <p class="text-muted-foreground">Все карточки из этой очереди пройдены.</p>
                    <Button variant="outline" class="mx-auto gap-2" onclick={() => navigate(`/${activeCardsetId}`)}>
                      <RotateCcw class="size-4" />
                      Вернуться к сессиям
                    </Button>
                  </div>
                {/if}
              </Card.Content>
            </div>
          </Card.Root>

          <Card.Root class="max-md:hidden">
            <Card.Header>
              <div class="flex items-center gap-2">
                <Keyboard class="size-4 text-muted-foreground" />
                <Card.Title>Клавиши</Card.Title>
              </div>
            </Card.Header>
            <Card.Content class="space-y-3 text-sm">
              <div class="flex items-center justify-between rounded-xl border px-3 py-2">
                <span>Показать ответ</span>
                <kbd class="rounded-md bg-muted px-2 py-1 text-xs">Space</kbd>
              </div>
              <div class="flex items-center justify-between rounded-xl border px-3 py-2">
                <span>Знаю</span>
                <kbd class="rounded-md bg-muted px-2 py-1 text-xs">Y</kbd>
              </div>
              <div class="flex items-center justify-between rounded-xl border px-3 py-2">
                <span>Пока не знаю</span>
                <kbd class="rounded-md bg-muted px-2 py-1 text-xs">N</kbd>
              </div>
            </Card.Content>
          </Card.Root>
        </div>

        {#if trainingState?.card}
          <div class="flex gap-3 max-md:hidden">
            <Button variant="outline" class="flex-1" onclick={toggleAnswer}>Показать / скрыть ответ</Button>
            <Button variant="destructive" class="flex-1" onclick={() => void markUnknown()} disabled={trainingLoading}>
              Пока не знаю
            </Button>
            <Button class="flex-1" onclick={() => void markKnown()} disabled={trainingLoading}>Знаю</Button>
          </div>
        {/if}

        {#if trainingState?.card && isMobile}
          <div class="fixed inset-x-0 bottom-0 border-t bg-background/95 p-4 backdrop-blur">
            <div class="mx-auto flex max-w-4xl items-center gap-3">
              <Button variant="outline" class="flex-1" onclick={toggleAnswer}>Показать / скрыть ответ</Button>
              <Badge variant="secondary" class="gap-1 px-3 py-2 text-xs">
                <GripHorizontal class="size-3.5" />
                вправо знаю, влево не знаю
              </Badge>
            </div>
          </div>
        {/if}
      </section>
    {/if}
  </div>
</main>
