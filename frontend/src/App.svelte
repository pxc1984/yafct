<script lang="ts">
  import promptTemplate from '../prompt.txt?raw'
  import {
    createCardSet,
    getCardSet,
    getSessionState,
    passCurrentCard,
    skipCurrentCard,
    startSession,
    uploadImage,
    type CardData,
    type CardImage,
    type CardSet,
    type CreateCardSetRequest,
    type SessionState,
  } from '$lib/api/flashcards'
  import HomePage from '$lib/components/app/home-page.svelte'
  import CardsetPage from '$lib/components/app/cardset-page.svelte'
  import SessionPage from '$lib/components/app/session-page.svelte'

  type SessionRecord = {
    id: string
    updatedAt: string
  }

  type RecentSession = SessionRecord & {
    cardsetId: string
    cardsetTitle: string
    cardsetAuthor: string
  }

  type StoredCardsetMeta = {
    title: string
    author: string
  }

  type Route =
    | { name: 'home' }
    | { name: 'cardset'; cardsetId: string }
    | { name: 'session'; cardsetId: string; sessionId: string }

  const STORAGE_KEY = 'flashcards-trainer-sessions'
  const CARDSET_META_STORAGE_KEY = 'flashcards-trainer-cardset-meta'
  const SWIPE_HINT_STORAGE_KEY = 'flashcards-trainer-swipe-hint-seen'
  const promptText = promptTemplate.trim()

  let route = $state<Route>(parseRoute(window.location.pathname))
  let sourceText = $state('')
  let setTitle = $state('')
  let setDescription = $state('')
  let setAuthor = $state('')
  let isCreating = $state(false)
  let createError = $state('')
  let createStatus = $state('')
  let copyState = $state<'idle' | 'done'>('idle')
  let cardsetLinkCopyState = $state<'idle' | 'done'>('idle')
  let sessionLinkCopyState = $state<'idle' | 'done'>('idle')
  let loadLinkError = $state('')

  let sessions = $state<SessionRecord[]>([])
  let recentSessions = $state<RecentSession[]>([])
  let sessionsLoading = $state(false)
  let sessionListError = $state('')
  let cardSetDetails = $state<CardSet | null>(null)
  let cardSetLoading = $state(false)
  let cardSetError = $state('')

  let trainingState = $state<SessionState | null>(null)
  let trainingLoading = $state(false)
  let trainingError = $state('')
  let isAnswerVisible = $state(false)
  let dragOffset = $state(0)
  let isDragging = $state(false)
  let dragStartX = 0
  let activePointerId: number | null = null
  let isMobile = $state(false)
  let showSwipeHint = $state(false)
  let uploadedImages = $state<Record<string, CardImage>>({})

  const progressValue = $derived.by(() => {
    if (!trainingState || trainingState.total === 0) {
      return 0
    }

    return Math.min(100, Math.round((trainingState.passed / trainingState.total) * 100))
  })

  const hasActiveCard = $derived(Boolean(trainingState?.card))
  const activeCardsetId = $derived(route.name === 'home' ? '' : route.cardsetId)
  const activeSessionId = $derived(route.name === 'session' ? route.sessionId : '')

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

  function buildAbsoluteUrl(path: string) {
    return new URL(path, window.location.origin).toString()
  }

  async function copyLink(path: string, target: 'cardset' | 'session') {
    await navigator.clipboard.writeText(buildAbsoluteUrl(path))

    if (target === 'cardset') {
      cardsetLinkCopyState = 'done'
    } else {
      sessionLinkCopyState = 'done'
    }

    window.setTimeout(() => {
      if (target === 'cardset') {
        cardsetLinkCopyState = 'idle'
        return
      }

      sessionLinkCopyState = 'idle'
    }, 1500)
  }

  function loadLink() {
    const nextLink = window.prompt('Вставь ссылку на набор или сессию.', buildAbsoluteUrl('/'))?.trim()

    if (!nextLink) {
      return
    }

    try {
      const pathname = new URL(nextLink, window.location.origin).pathname
      navigate(pathname)
      loadLinkError = ''
    } catch {
      loadLinkError = 'Не удалось распознать ссылку.'
    }
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

  function readStoredCardsetMeta() {
    if (typeof localStorage === 'undefined') {
      return {}
    }

    const raw = localStorage.getItem(CARDSET_META_STORAGE_KEY)

    if (!raw) {
      return {}
    }

    try {
      return JSON.parse(raw) as Record<string, StoredCardsetMeta>
    } catch {
      return {}
    }
  }

  function storeCardsetMeta(cardsetId: string, meta: StoredCardsetMeta) {
    if (typeof localStorage === 'undefined') {
      return
    }

    const parsed = readStoredCardsetMeta()
    parsed[cardsetId] = meta
    localStorage.setItem(CARDSET_META_STORAGE_KEY, JSON.stringify(parsed))
  }

  function syncRecentSessions() {
    if (typeof localStorage === 'undefined') {
      recentSessions = []
      return
    }

    const raw = localStorage.getItem(STORAGE_KEY)

    if (!raw) {
      recentSessions = []
      return
    }

    try {
      const cardsetMeta = readStoredCardsetMeta()
      const parsed = JSON.parse(raw) as Record<string, SessionRecord[]>
      recentSessions = Object.entries(parsed)
        .flatMap(([cardsetId, cardsetSessions]) =>
          (Array.isArray(cardsetSessions) ? cardsetSessions : []).map((session) => ({
            ...session,
            cardsetId,
            cardsetTitle: cardsetMeta[cardsetId]?.title || cardsetId,
            cardsetAuthor: cardsetMeta[cardsetId]?.author || '',
          })),
        )
        .sort((left, right) => new Date(right.updatedAt).getTime() - new Date(left.updatedAt).getTime())
        .slice(0, 8)
    } catch {
      recentSessions = []
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
    syncRecentSessions()
  }

  function syncSessions(cardsetId: string) {
    sessions = readStoredSessions(cardsetId)
  }

  function removeSession(cardsetId: string, sessionId: string) {
    if (typeof localStorage === 'undefined') {
      return
    }

    const raw = localStorage.getItem(STORAGE_KEY)
    const parsed = raw ? (JSON.parse(raw) as Record<string, SessionRecord[]>) : {}
    const current = Array.isArray(parsed[cardsetId]) ? parsed[cardsetId] : []

    parsed[cardsetId] = current.filter((session) => session.id !== sessionId)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(parsed))
    syncSessions(cardsetId)
    syncRecentSessions()
  }

  function syncSwipeHint() {
    if (typeof localStorage === 'undefined') {
      showSwipeHint = false
      return
    }

    const isSeen = localStorage.getItem(SWIPE_HINT_STORAGE_KEY) === '1'
    showSwipeHint = !isSeen

    if (!isSeen) {
      localStorage.setItem(SWIPE_HINT_STORAGE_KEY, '1')
    }
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
          questionImages.push(resolveUploadedImage(readFieldValue(line, 'QUESTION_IMAGE::'), index + 1))
          continue
        }

        if (line.startsWith('ANSWER_IMAGE::')) {
          answerImages.push(resolveUploadedImage(readFieldValue(line, 'ANSWER_IMAGE::'), index + 1))
          continue
        }

        throw new Error(`Карточка ${index + 1} содержит строку в неверном формате.`)
      }

      if (!question || !answer) {
        throw new Error(`У карточки ${index + 1} обязательны QUESTION и ANSWER.`)
      }

      if (questionImages.length > 5 || answerImages.length > 5) {
        throw new Error(`У карточки ${index + 1} можно добавить не больше 5 изображений к вопросу или ответу.`)
      }

      return { question, answer, remarks, questionImages, answerImages } satisfies CardData
    })
  }

  function resolveUploadedImage(imageId: string, cardNumber: number) {
    const image = uploadedImages[imageId]

    if (!image) {
      throw new Error(`Карточка ${cardNumber} содержит изображение ${imageId}, которого нет в текущем черновике.`)
    }

    return image
  }

  async function handleUploadImage(file: File) {
    const image = await uploadImage(file)
    uploadedImages = { ...uploadedImages, [image.id]: image }
    return image
  }

  async function copyPrompt() {
    await navigator.clipboard.writeText(promptText)
    copyState = 'done'

    window.setTimeout(() => {
      copyState = 'idle'
    }, 1500)
  }

  async function createSet(nextCards?: CardData[]) {
    createError = ''
    createStatus = 'Подготавливаю карточки...'

    let cards: CardData[]

    try {
      cards = nextCards ?? parseCardData(sourceText)
    } catch (error) {
      createError = error instanceof Error ? error.message : 'Не удалось разобрать карточки.'
      createStatus = ''
      return
    }

    isCreating = true

    try {
      createStatus = 'Собираю запрос...'

      const payload: CreateCardSetRequest = {
        title: setTitle,
        description: setDescription,
        author: setAuthor,
        cards,
      }

      createStatus = 'Отправляю набор на сервер...'
      const { id } = await createCardSet(payload)
      createStatus = 'Открываю созданный набор...'
      navigate(`/${id}`)
      storeCardsetMeta(id, {
        title: setTitle.trim() || id,
        author: setAuthor.trim(),
      })
      syncRecentSessions()
    } catch (error) {
      createError = error instanceof Error ? error.message : 'Не удалось создать набор.'
      createStatus = 'Создание остановилось с ошибкой.'
    } finally {
      isCreating = false

      if (!createError) {
        createStatus = ''
      }
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

  async function loadCardSet(cardsetId: string) {
    cardSetLoading = true
    cardSetError = ''

    try {
      cardSetDetails = await getCardSet(cardsetId)
      storeCardsetMeta(cardsetId, {
        title: cardSetDetails.title || cardsetId,
        author: cardSetDetails.author || '',
      })
      syncRecentSessions()
    } catch (error) {
      cardSetError = error instanceof Error ? error.message : 'Не удалось загрузить набор.'
    } finally {
      cardSetLoading = false
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

    if (activePointerId !== null) {
      return
    }

    event.preventDefault()
    event.currentTarget instanceof HTMLElement && event.currentTarget.setPointerCapture(event.pointerId)

    activePointerId = event.pointerId
    isDragging = true
    dragStartX = event.clientX
  }

  function handlePointerMove(event: PointerEvent) {
    if (!isDragging || event.pointerId !== activePointerId) {
      return
    }

    event.preventDefault()
    dragOffset = event.clientX - dragStartX
  }

  async function handlePointerUp(event: PointerEvent) {
    if (!isDragging || event.pointerId !== activePointerId) {
      return
    }

    event.preventDefault()

    if (event.currentTarget instanceof HTMLElement && event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }

    isDragging = false
    activePointerId = null

    if (dragOffset >= 96) {
      dragOffset = 0
      await markKnown()
      return
    }

    if (dragOffset <= -96) {
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
    if (route.name === 'home') {
      syncRecentSessions()
    }

    if (route.name === 'cardset') {
      syncSessions(route.cardsetId)
      void loadCardSet(route.cardsetId)
    }

    if (route.name === 'session') {
      syncSwipeHint()
      cardSetDetails = null
      void loadTraining(route.cardsetId, route.sessionId)
    }
  })

  $effect(() => {
    if (!isDragging || activePointerId === null) {
      return
    }

    const onWindowPointerMove = (event: PointerEvent) => {
      void handlePointerMove(event)
    }

    const onWindowPointerUp = (event: PointerEvent) => {
      void handlePointerUp(event)
    }

    window.addEventListener('pointermove', onWindowPointerMove, { passive: false })
    window.addEventListener('pointerup', onWindowPointerUp, { passive: false })
    window.addEventListener('pointercancel', onWindowPointerUp, { passive: false })

    return () => {
      window.removeEventListener('pointermove', onWindowPointerMove)
      window.removeEventListener('pointerup', onWindowPointerUp)
      window.removeEventListener('pointercancel', onWindowPointerUp)
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

  $effect(() => {
    if (!isMobile || !isDragging) {
      return
    }

    const previousOverflow = document.body.style.overflow
    const previousTouchAction = document.body.style.touchAction

    document.body.style.overflow = 'hidden'
    document.body.style.touchAction = 'none'

    return () => {
      document.body.style.overflow = previousOverflow
      document.body.style.touchAction = previousTouchAction
    }
  })
</script>

<main class="min-h-screen bg-background text-foreground">
  <div class="mx-auto flex min-h-screen w-full max-w-5xl flex-col px-4 py-6 sm:px-6 sm:py-8">
    {#if route.name === 'home'}
      <HomePage
        promptText={promptText}
        bind:sourceText
        bind:setTitle
        bind:setDescription
        bind:setAuthor
        {parseCardData}
        resolveImageById={(imageId) => uploadedImages[imageId] ?? null}
        {recentSessions}
        isCreating={isCreating}
        createError={createError}
        {createStatus}
        copyState={copyState}
        {loadLinkError}
        {formatDate}
        onCopyPrompt={copyPrompt}
        onLoadLink={loadLink}
        onUploadImage={handleUploadImage}
        onCreateSet={createSet}
        onNavigate={navigate}
      />
    {/if}

    {#if route.name === 'cardset'}
      <CardsetPage
        cardsetId={route.cardsetId}
        cardSetDetails={cardSetDetails}
        cardSetLoading={cardSetLoading}
        cardSetError={cardSetError}
        sessions={sessions}
        sessionListError={sessionListError}
        sessionsLoading={sessionsLoading}
        {isMobile}
        copyLinkState={cardsetLinkCopyState}
        onCreateSession={() => createSession(activeCardsetId)}
        onCopyLink={() => void copyLink(`/${activeCardsetId}`, 'cardset')}
        onNavigate={navigate}
        onRemoveSession={(sessionId) => removeSession(activeCardsetId, sessionId)}
        {formatDate}
      />
    {/if}

    {#if route.name === 'session'}
      <SessionPage
        cardsetId={route.cardsetId}
        {trainingState}
        {trainingLoading}
        {trainingError}
        {isAnswerVisible}
        {isDragging}
        {dragOffset}
        {progressValue}
        {isMobile}
        {showSwipeHint}
        copyLinkState={sessionLinkCopyState}
        onNavigate={navigate}
        onCopyLink={() => void copyLink(`/${activeCardsetId}/${activeSessionId}`, 'session')}
        onToggleAnswer={toggleAnswer}
        onMarkKnown={() => void markKnown()}
        onMarkUnknown={() => void markUnknown()}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={(event) => void handlePointerUp(event)}
      />
    {/if}
  </div>
</main>
