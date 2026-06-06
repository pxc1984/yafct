<script lang="ts">
  import Copy from '@lucide/svelte/icons/copy'

  import type { CardData } from '$lib/api/flashcards'

  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'

  let {
    promptText,
    sourceText = $bindable(''),
    setTitle = $bindable(''),
    setDescription = $bindable(''),
    setAuthor = $bindable(''),
    parseCardData,
    isCreating,
    createError,
    copyState,
    onCopyPrompt,
    onCreateSet,
  }: {
    promptText: string
    sourceText: string
    setTitle: string
    setDescription: string
    setAuthor: string
    parseCardData: (input: string) => CardData[]
    isCreating: boolean
    createError: string
    copyState: 'idle' | 'done'
    onCopyPrompt: () => void | Promise<void>
    onCreateSet: () => void | Promise<void>
  } = $props()

  const sourcePlaceholder = `QUESTION:: Что такое closure?
ANSWER:: Функция вместе с лексическим окружением.
REMARK:: Удобно для инкапсуляции состояния.

QUESTION:: Что возвращает выражение $2^3$?
ANSWER:: $8$
REMARK:: `

  let previewMode = $state<'text' | 'list'>('text')
  let previewCards = $state<CardData[]>([])
  let previewError = $state('')
  let syncedPreviewSourceText = ''

  function formatCardData(cards: CardData[]) {
    return cards
      .map(
        (card) =>
          `QUESTION:: ${card.question}\nANSWER:: ${card.answer}\nREMARK:: ${card.remarks}`,
      )
      .join('\n\n')
  }

  function syncPreviewFromSource() {
    try {
      previewCards = parseCardData(sourceText)
      previewError = ''
    } catch (error) {
      previewCards = []
      previewError = error instanceof Error ? error.message : 'Не удалось разобрать карточки.'
    }

    syncedPreviewSourceText = sourceText
  }

  function setPreviewMode(mode: 'text' | 'list') {
    previewMode = mode

    if (mode === 'list' && syncedPreviewSourceText !== sourceText) {
      syncPreviewFromSource()
    }
  }

  function updateCardField(index: number, field: keyof CardData, value: string) {
    previewCards = previewCards.map((card, cardIndex) =>
      cardIndex === index ? { ...card, [field]: value } : card,
    )

    const nextSourceText = formatCardData(previewCards)
    syncedPreviewSourceText = nextSourceText
    sourceText = nextSourceText
    previewError = ''
  }

  $effect(() => {
    if (previewMode === 'list' && syncedPreviewSourceText !== sourceText) {
      syncPreviewFromSource()
    }
  })
</script>

<section class="mx-auto flex w-full flex-1 items-center">
  <Card.Root class="border-border/70 bg-card/85 shadow-sm backdrop-blur">
    <Card.Header class="gap-4">
      <div class="space-y-2">
        <Card.Title class="text-3xl sm:text-4xl">Тренировка по карточкам</Card.Title>
        <Card.Description>
          Если лень писать самостоятельно карточки можешь воспользоваться вот этим системным промптом.
        </Card.Description>
      </div>
    </Card.Header>
    <Card.Content class="space-y-4">
      <div class="relative rounded-2xl border bg-background/70 p-4 pr-16">
        <Button
          variant="outline"
          size="icon-sm"
          class="absolute top-3 right-3"
          onclick={onCopyPrompt}
          aria-label="Скопировать промпт"
        >
          <Copy class="size-4" />
        </Button>
        <pre class="max-h-[4.75rem] overflow-auto whitespace-pre-wrap text-sm text-muted-foreground">{promptText}</pre>
      </div>

      <div class="grid gap-4 lg:grid-cols-2 lg:items-stretch">
        <div class="flex min-h-[28rem] flex-col gap-4">
          <label class="space-y-2">
            <span class="text-sm font-medium">Название набора</span>
            <input
              bind:value={setTitle}
              maxlength="120"
              class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-12 w-full rounded-2xl border bg-transparent px-4 py-3 text-sm outline-none focus-visible:ring-3"
              placeholder="Математический анализ к коллоквиуму"
            />
          </label>

          <label class="flex flex-1 flex-col gap-2">
            <span class="text-sm font-medium">Описание</span>
            <textarea
              bind:value={setDescription}
              class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 min-h-0 flex-1 resize-none overflow-auto rounded-2xl border bg-transparent px-4 py-3 text-sm outline-none focus-visible:ring-3"
              placeholder="Сделано без любви и с помощью ллм, зато работает."
            ></textarea>
          </label>

          <label class="space-y-2">
            <span class="text-sm font-medium">Имя автора</span>
            <input
              bind:value={setAuthor}
              class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-12 w-full rounded-2xl border bg-transparent px-4 py-3 text-sm outline-none focus-visible:ring-3"
              placeholder="Игорь <@igamamaev>"
            />
          </label>
        </div>

        <div class="flex h-[18rem] flex-col gap-2">
          <div class="flex items-center justify-between gap-3">
            <span class="text-sm font-medium">Текст с карточками</span>
            <Button
              variant="outline"
              size="sm"
              onclick={() => setPreviewMode(previewMode === 'text' ? 'list' : 'text')}
            >
              {previewMode === 'text' ? 'Показать список' : 'Показать текст'}
            </Button>
          </div>

          {#if previewMode === 'text'}
            <textarea
              bind:value={sourceText}
              class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 min-h-0 h-[28rem] flex-1 resize-none overflow-auto rounded-2xl border bg-transparent px-4 py-3 text-sm outline-none focus-visible:ring-3"
              placeholder={sourcePlaceholder}
            ></textarea>
          {:else}
            <div
              data-testid="cards-list-container"
              class="dark:bg-input/30 border-input min-h-0 h-[18rem] flex-1 overflow-hidden rounded-2xl border bg-transparent text-sm"
            >
              {#if previewError}
                <div class="h-full overflow-auto px-4 py-3">
                  <p class="text-destructive">Не удалось распарсить текст: {previewError}</p>
                </div>
              {:else}
                <div class="h-full overflow-auto px-4 py-3">
                  <div class="space-y-4">
                  {#each previewCards as card, index (index)}
                    <section data-testid={`preview-card-${index}`} class="space-y-2 rounded-xl border border-border/70 bg-background/70 p-3">
                      <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Карточка {index + 1}</p>
                      <label class="flex flex-col gap-1">
                        <span class="font-medium">Вопрос</span>
                        <textarea
                          value={card.question}
                          class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-15 overflow-auto resize-y rounded-xl border bg-transparent px-3 py-2 outline-none focus-visible:ring-3"
                          oninput={(event) => updateCardField(index, 'question', event.currentTarget.value)}
                        ></textarea>
                      </label>
                      <label class="flex flex-col gap-1">
                        <span class="font-medium">Ответ</span>
                        <textarea
                          value={card.answer}
                          class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-15 overflow-auto resize-y rounded-xl border bg-transparent px-3 py-2 outline-none focus-visible:ring-3"
                          oninput={(event) => updateCardField(index, 'answer', event.currentTarget.value)}
                        ></textarea>
                      </label>
                      <label class="flex flex-col gap-1">
                        <span class="font-medium">Ремарка</span>
                        <textarea
                          value={card.remarks}
                          class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-10 overflow-auto resize-y rounded-xl border bg-transparent px-3 py-2 outline-none focus-visible:ring-3"
                          oninput={(event) => updateCardField(index, 'remarks', event.currentTarget.value)}
                        ></textarea>
                      </label>
                    </section>
                  {/each}
                  </div>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      </div>

      {#if createError}
        <p class="text-sm text-destructive">{createError}</p>
      {/if}
    </Card.Content>
    <Card.Footer class="justify-between gap-3 max-sm:flex-col max-sm:items-stretch">
      <p class="text-sm text-muted-foreground">
        {copyState === 'done' ? 'Промпт скопирован.' : ''}
      </p>
      <Button size="lg" onclick={onCreateSet} disabled={isCreating || !sourceText.trim()}>
        {isCreating ? 'Создание...' : 'Создать набор'}
      </Button>
    </Card.Footer>
  </Card.Root>
</section>
