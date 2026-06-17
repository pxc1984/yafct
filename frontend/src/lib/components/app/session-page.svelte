<script lang="ts">
  import X from '@lucide/svelte/icons/x'
  import Panzoom from '@panzoom/panzoom'
  import type { PanzoomObject } from '@panzoom/panzoom'
  import Copy from '@lucide/svelte/icons/copy'
  import Keyboard from '@lucide/svelte/icons/keyboard'
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw'

  import type { CardImage, SessionState } from '$lib/api/flashcards'
  import RichMathText from '$lib/components/rich-math-text.svelte'
  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'

  let {
    cardsetId,
    trainingState,
    trainingLoading,
    trainingError,
    isAnswerVisible,
    isDragging,
    dragOffset,
    progressValue,
    isMobile,
    showSwipeHint,
    onNavigate,
    onCopyLink,
    onToggleAnswer,
    onMarkKnown,
    onMarkUnknown,
    onPointerDown,
  }: {
    cardsetId: string
    trainingState: SessionState | null
    trainingLoading: boolean
    trainingError: string
    isAnswerVisible: boolean
    isDragging: boolean
    dragOffset: number
    progressValue: number
    isMobile: boolean
    showSwipeHint: boolean
    onNavigate: (path: string) => void
    onCopyLink: () => void | Promise<void>
    onToggleAnswer: () => void
    onMarkKnown: () => void | Promise<void>
    onMarkUnknown: () => void | Promise<void>
    onPointerDown: (event: PointerEvent) => void
  } = $props()

  let fullscreenImage = $state<CardImage | null>(null)
  let fullscreenViewport = $state<HTMLDivElement | null>(null)
  let fullscreenImageElement = $state<HTMLImageElement | null>(null)
  let fullscreenPanzoom: PanzoomObject | null = null

  function openFullscreenImage(image: CardImage) {
    fullscreenImage = image
  }

  function closeFullscreenImage() {
    fullscreenImage = null
  }

  function handleQuestionImageClick(event: MouseEvent, image: CardImage) {
    event.stopPropagation()
    openFullscreenImage(image)
  }

  function handleAnswerImageClick(event: MouseEvent, image: CardImage) {
    if (!isAnswerVisible) {
      return
    }

    event.stopPropagation()
    openFullscreenImage(image)
  }

  function handleWindowKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && fullscreenImage) {
      closeFullscreenImage()
    }
  }

  function handleFullscreenBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      closeFullscreenImage()
    }
  }

  $effect(() => {
    if (!fullscreenImage || !fullscreenViewport || !fullscreenImageElement) {
      return
    }

    const viewport = fullscreenViewport
    const imageElement = fullscreenImageElement
    const panzoom = Panzoom(imageElement, {
      contain: 'outside',
      cursor: 'grab',
      maxScale: 5,
      minScale: 1,
      step: 0.25,
    })

    const handleWheel = (event: WheelEvent) => {
      event.preventDefault()
      panzoom.zoomWithWheel(event)
    }

    viewport.addEventListener('wheel', handleWheel, { passive: false })
    fullscreenPanzoom = panzoom

    return () => {
      viewport.removeEventListener('wheel', handleWheel)
      panzoom.destroy()
      panzoom.resetStyle()
      if (fullscreenPanzoom === panzoom) {
        fullscreenPanzoom = null
      }
    }
  })

</script>

<svelte:window onkeydown={handleWindowKeydown} />

<section class="mx-auto flex w-full max-w-8xl flex-1 flex-col gap-4 overflow-x-hidden pb-28 sm:pb-8">
  <div class="flex items-center justify-between gap-3">
    <div>
      <p class="text-sm text-muted-foreground">{trainingState?.author ? `Автор: ${trainingState.author}` : `${cardsetId}`}</p>
      <h1 class="text-2xl font-semibold">{trainingState?.title || 'Тренировка'}</h1>
    </div>
    <div class="flex items-center gap-2">
      {#if isMobile}
        <Button variant="outline" onclick={onCopyLink}>
          <Copy class="size-4" />
        </Button>
      {/if}
    </div>
  </div>

  <Card.Root size="sm">
    <Card.Content class="space-y-2 pb-3">
      <div class="flex items-center justify-between text-sm text-muted-foreground">
        <span>Прогресс</span>
        <span>{trainingState?.passed ?? 0} / {trainingState?.total ?? 0}</span>
      </div>
      <div class="h-2 overflow-hidden rounded-full bg-muted">
        <div class="h-full rounded-full bg-emerald-500 transition-all" style={`width: ${progressValue}%`}></div>
      </div>
    </Card.Content>
  </Card.Root>

  <div class="grid gap-4 md:grid-cols-[1fr_220px]">
    <Card.Root
      class={`select-none overflow-hidden transition-transform ${isMobile ? (isDragging ? '' : 'duration-200') : 'duration-200'} ${isMobile ? (isDragging ? 'cursor-grabbing' : 'cursor-grab') : ''}`}
      style={isMobile ? `transform: translate3d(${dragOffset}px, 0, 0) rotate(${Math.max(-10, Math.min(10, dragOffset / 24))}deg);` : undefined}
      data-dragging={isDragging}
      data-swipe-card
      onpointerdown={isMobile ? onPointerDown : undefined}
    >
      <div class="h-full">
        <Card.Content class="space-y-6">
          {#if trainingLoading && !trainingState}
            <p class="text-muted-foreground">Загрузка...</p>
          {:else if trainingError}
            <p class="text-destructive">{trainingError}</p>
          {:else if trainingState?.card}
            <div class="space-y-3">
              <RichMathText text={trainingState.card.question} class="text-2xl leading-tight font-semibold sm:text-3xl" />
              {#if trainingState.card.questionImages.length > 0}
                <div class={trainingState.card.questionImages.length === 1 ? 'space-y-3' : 'grid gap-3 sm:grid-cols-2'}>
                  {#each trainingState.card.questionImages as image (image.id)}
                    <button
                      type="button"
                      class="block w-full cursor-zoom-in rounded-2xl"
                      onpointerdown={(event) => event.stopPropagation()}
                      onclick={(event) => handleQuestionImageClick(event, image)}
                    >
                      <img
                        src={`data:${image.mimeType};base64,${image.dataBase64}`}
                        alt="Изображение вопроса"
                        class={`w-full rounded-2xl border bg-background/60 object-contain ${trainingState.card.questionImages.length === 1 ? 'h-auto max-h-none' : 'max-h-72'}`}
                        draggable="false"
                      />
                    </button>
                  {/each}
                </div>
              {/if}
              {#if trainingState.card.remarks}
                <RichMathText text={trainingState.card.remarks} class="text-sm text-muted-foreground" />
              {/if}
            </div>

            <div class="rounded-2xl border bg-background/60 p-4">
              <p class="mb-2 text-sm font-medium text-muted-foreground">Ответ</p>
              <div
                class="group relative cursor-pointer overflow-hidden rounded-lg"
                onclick={onToggleAnswer}
                role="button"
                tabindex="0"
                onkeydown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    onToggleAnswer()
                  }
                }}
              >
                {#if !isAnswerVisible}
                  <div class="pointer-events-none absolute inset-0 z-10 rounded-lg bg-gradient-to-br from-muted/20 via-background/10 to-muted/20 opacity-80 transition-opacity duration-500">
                    <div class="shimmer size-full"></div>
                  </div>
                {/if}
                <div class="fog-text" class:revealed={isAnswerVisible}>
                  <RichMathText text={trainingState.card.answer} class="text-lg leading-relaxed" />
                  {#if trainingState.card.answerImages.length > 0}
                    <div class={`mt-4 ${trainingState.card.answerImages.length === 1 ? 'space-y-3' : 'grid gap-3 sm:grid-cols-2'}`}>
                      {#each trainingState.card.answerImages as image (image.id)}
                        {#if isAnswerVisible}
                          <button
                            type="button"
                            class="block w-full cursor-zoom-in rounded-2xl"
                            onpointerdown={(event) => event.stopPropagation()}
                            onclick={(event) => handleAnswerImageClick(event, image)}
                          >
                            <img
                              src={`data:${image.mimeType};base64,${image.dataBase64}`}
                              alt="Изображение ответа"
                              class={`w-full rounded-2xl border bg-background/60 object-contain ${trainingState.card.answerImages.length === 1 ? 'h-auto max-h-none' : 'max-h-72'}`}
                              draggable="false"
                            />
                          </button>
                        {:else}
                          <img
                            src={`data:${image.mimeType};base64,${image.dataBase64}`}
                            alt="Изображение ответа"
                            class={`w-full rounded-2xl border bg-background/60 object-contain ${trainingState.card.answerImages.length === 1 ? 'h-auto max-h-none' : 'max-h-72'}`}
                            draggable="false"
                          />
                        {/if}
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>
            </div>
          {:else}
            <div class="space-y-3 py-8 text-center">
              <p class="text-2xl font-semibold">Сессия завершена</p>
              <p class="text-muted-foreground">Все карточки из этой очереди пройдены.</p>
              <Button variant="outline" class="mx-auto gap-2" onclick={() => onNavigate(`/${cardsetId}`)}>
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
      <Button variant="outline" class="flex-1" onclick={onToggleAnswer}>Показать / скрыть ответ</Button>
      <Button variant="destructive" class="flex-1" onclick={onMarkUnknown} disabled={trainingLoading}>
        Пока не знаю
      </Button>
      <Button class="flex-1" onclick={onMarkKnown} disabled={trainingLoading}>Знаю</Button>
    </div>
  {/if}

  {#if trainingState?.card && isMobile && showSwipeHint}
    <div class="fixed inset-x-0 bottom-0 border-t bg-background/95 p-4 backdrop-blur flex flex-row gap-2">
      <Button variant="outline" class="flex-1">&lt;- Не знаю</Button>
      <Button variant="outline" class="flex-1">Знаю -&gt;</Button>
    </div>
  {/if}

  {#if fullscreenImage}
    <div
      class="fixed inset-0 z-50 flex bg-background/95 backdrop-blur-sm"
      role="button"
      tabindex="0"
      aria-label="Закрыть полноэкранное изображение"
      onclick={handleFullscreenBackdropClick}
      onkeydown={(event) => {
        if (event.key === 'Enter' || event.key === ' ' || event.key === 'Escape') {
          event.preventDefault()
          closeFullscreenImage()
        }
      }}
    >
      <Button
        variant="outline"
        size="icon-sm"
        class="absolute top-4 left-4 z-60"
        aria-label="Закрыть изображение"
        onclick={closeFullscreenImage}
      >
        <X class="size-4" />
      </Button>
      <div bind:this={fullscreenViewport} class="flex h-full w-full overflow-hidden pt-12">
        <img
          bind:this={fullscreenImageElement}
          src={`data:${fullscreenImage.mimeType};base64,${fullscreenImage.dataBase64}`}
          alt="Полноэкранное изображение"
          class="max-h-full max-w-full touch-none object-contain"
          draggable="false"
        />
      </div>
    </div>
  {/if}
</section>

<style>
  .fog-text {
    filter: blur(6px);
    transition: filter 0.5s ease;
  }
  .fog-text.revealed {
    filter: blur(0);
  }
  .shimmer {
    position: relative;
    overflow: hidden;
  }
  .shimmer::after {
    content: '';
    position: absolute;
    inset: -50%;
    background: linear-gradient(
      135deg,
      transparent 30%,
      rgba(255, 255, 255, 0.06) 50%,
      transparent 70%
    );
    background-size: 200% 200%;
    animation: fog-shimmer 4s ease-in-out infinite;
  }
  @keyframes fog-shimmer {
    0%, 100% { background-position: 0% 0%; }
    50% { background-position: 100% 100%; }
  }
</style>
