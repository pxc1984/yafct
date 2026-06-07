<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left'
  import Copy from '@lucide/svelte/icons/copy'
  import Keyboard from '@lucide/svelte/icons/keyboard'
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw'

  import type { SessionState } from '$lib/api/flashcards'
  import RichMathText from '$lib/components/rich-math-text.svelte'
  import { Badge } from '$lib/components/ui/badge'
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
    copyLinkState,
    onNavigate,
    onCopyLink,
    onToggleAnswer,
    onMarkKnown,
    onMarkUnknown,
    onPointerDown,
    onPointerMove,
    onPointerUp,
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
    copyLinkState: 'idle' | 'done'
    onNavigate: (path: string) => void
    onCopyLink: () => void | Promise<void>
    onToggleAnswer: () => void
    onMarkKnown: () => void | Promise<void>
    onMarkUnknown: () => void | Promise<void>
    onPointerDown: (event: PointerEvent) => void
    onPointerMove: (event: PointerEvent) => void
    onPointerUp: (event: PointerEvent) => void | Promise<void>
  } = $props()
</script>

<section class="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-4 overflow-x-hidden pb-28 sm:pb-8">
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
      <Button variant="outline" size="icon-sm" onclick={() => onNavigate(`/${cardsetId}`)} aria-label="К сессиям" title="К сессиям">
        <ArrowLeft class="size-4" />
      </Button>
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
      class={`select-none overflow-hidden transition-transform ${isDragging ? '' : 'duration-200'} ${isDragging ? 'cursor-grabbing' : 'cursor-grab'}`}
      style={`transform: translate3d(${dragOffset}px, 0, 0) rotate(${Math.max(-10, Math.min(10, dragOffset / 24))}deg);`}
      data-dragging={isDragging}
      data-swipe-card
      onpointerdown={onPointerDown}
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
                <div class="grid gap-3 sm:grid-cols-2">
                  {#each trainingState.card.questionImages as image (image.id)}
                    <img src={`data:${image.mimeType};base64,${image.dataBase64}`} alt="Изображение вопроса" class="pointer-events-none max-h-72 w-full rounded-2xl border bg-background/60 object-contain" draggable="false" />
                  {/each}
                </div>
              {/if}
              {#if trainingState.card.remarks}
                <RichMathText text={trainingState.card.remarks} class="text-sm text-muted-foreground" />
              {/if}
            </div>

            <div class="rounded-2xl border bg-background/60 p-4">
              <p class="mb-2 text-sm font-medium text-muted-foreground">Ответ</p>
              {#if isAnswerVisible}
                <RichMathText text={trainingState.card.answer} class="text-lg leading-relaxed" />
                {#if trainingState.card.answerImages.length > 0}
                  <div class="mt-4 grid gap-3 sm:grid-cols-2">
                    {#each trainingState.card.answerImages as image (image.id)}
                      <img src={`data:${image.mimeType};base64,${image.dataBase64}`} alt="Изображение ответа" class="pointer-events-none max-h-72 w-full rounded-2xl border bg-background/60 object-contain" draggable="false" />
                    {/each}
                  </div>
                {/if}
              {:else}
                <p class="text-lg text-muted-foreground">Нажми показать ответ или пробел.</p>
              {/if}
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
    <div class="fixed inset-x-0 bottom-0 border-t bg-background/95 p-4 backdrop-blur">
      <Button variant="outline" class="flex-1" onclick={onToggleAnswer}>Показать / скрыть ответ</Button>
    </div>
  {:else if trainingState?.card && isMobile}
    <div class="fixed inset-x-0 bottom-0 border-t bg-background/95 p-4 backdrop-blur">
      <div class="mx-auto flex max-w-4xl items-center gap-3">
        <Button variant="outline" class="flex-1" onclick={onToggleAnswer}>Показать / скрыть ответ</Button>
      </div>
    </div>
  {/if}
</section>
