<script lang="ts">
  import Copy from '@lucide/svelte/icons/copy'
  import ImagePlus from '@lucide/svelte/icons/image-plus'
  import Plus from '@lucide/svelte/icons/plus'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import X from '@lucide/svelte/icons/x'

  import type { CardData, CardImage } from '$lib/api/flashcards'

  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'

  let {
    promptText,
    sourceText = $bindable(''),
    setTitle = $bindable(''),
    setDescription = $bindable(''),
    setAuthor = $bindable(''),
    parseCardData,
    resolveImageById,
    isCreating,
    createError,
    createStatus,
    copyState,
    loadLinkError,
    onCopyPrompt,
    onLoadLink,
    onUploadImage,
    onCreateSet,
  }: {
    promptText: string
    sourceText: string
    setTitle: string
    setDescription: string
    setAuthor: string
    parseCardData: (input: string) => CardData[]
    resolveImageById: (imageId: string) => CardImage | null
    isCreating: boolean
    createError: string
    createStatus: string
    copyState: 'idle' | 'done'
    loadLinkError: string
    onCopyPrompt: () => void | Promise<void>
    onLoadLink: () => void
    onUploadImage: (file: File) => Promise<CardImage>
    onCreateSet: (cards?: CardData[]) => void | Promise<void>
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
  let pendingDeleteIndex = $state<number | null>(null)
  let deleteConfirmationTimeout: ReturnType<typeof setTimeout> | null = null
  let uploadError = $state('')
  let uploadTarget = $state<{ index: number; field: 'questionImages' | 'answerImages' } | null>(null)
  let fileInput: HTMLInputElement | null = null

  function formatCardData(cards: CardData[]) {
    return cards
      .map(
        (card) =>
          [
            `QUESTION:: ${card.question}`,
            ...card.questionImages.map((image) => `QUESTION_IMAGE:: ${image.id}`),
            `ANSWER:: ${card.answer}`,
            ...card.answerImages.map((image) => `ANSWER_IMAGE:: ${image.id}`),
            `REMARK:: ${card.remarks}`,
          ].join('\n'),
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
    uploadError = ''
  }

  function clearPendingDelete() {
    pendingDeleteIndex = null

    if (deleteConfirmationTimeout) {
      clearTimeout(deleteConfirmationTimeout)
      deleteConfirmationTimeout = null
    }
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

  function addPreviewCard() {
    clearPendingDelete()

    const nextCardNumber = previewCards.length + 1
    previewCards = [
      ...previewCards,
      {
        question: `Новый вопрос ${nextCardNumber}`,
        answer: `Новый ответ ${nextCardNumber}`,
        remarks: '',
        questionImages: [],
        answerImages: [],
      },
    ]

    const nextSourceText = formatCardData(previewCards)
    syncedPreviewSourceText = nextSourceText
    sourceText = nextSourceText
    previewError = ''
  }

  function requestDeleteCard(index: number) {
    if (pendingDeleteIndex === index) {
      previewCards = previewCards.filter((_, cardIndex) => cardIndex !== index)

      const nextSourceText = formatCardData(previewCards)
      syncedPreviewSourceText = nextSourceText
      sourceText = nextSourceText
      previewError = ''
      clearPendingDelete()
      return
    }

    clearPendingDelete()
    pendingDeleteIndex = index
    deleteConfirmationTimeout = setTimeout(() => {
      pendingDeleteIndex = null
      deleteConfirmationTimeout = null
    }, 4000)
  }

  function updateCardImages(index: number, field: 'questionImages' | 'answerImages', images: CardImage[]) {
    previewCards = previewCards.map((card, cardIndex) =>
      cardIndex === index ? { ...card, [field]: images } : card,
    )

    const nextSourceText = formatCardData(previewCards)
    syncedPreviewSourceText = nextSourceText
    sourceText = nextSourceText
    previewError = ''
  }

  function openImagePicker(index: number, field: 'questionImages' | 'answerImages') {
    uploadTarget = { index, field }
    fileInput?.click()
  }

  async function uploadCardImage(index: number, field: 'questionImages' | 'answerImages', file: File) {
    const card = previewCards[index]
    if (!card) {
      return
    }

    if (card[field].length >= 5) {
      uploadError = 'К вопросу или ответу можно прикрепить не больше 5 изображений.'
      return
    }

    uploadError = ''
    const image = await onUploadImage(file)
    updateCardImages(index, field, [...card[field], image])
  }

  function removeCardImage(index: number, field: 'questionImages' | 'answerImages', imageId: string) {
    const card = previewCards[index]
    if (!card) {
      return
    }

    updateCardImages(index, field, card[field].filter((image) => image.id !== imageId))
  }

  async function handleFileInputChange(event: Event) {
    const input = event.currentTarget as HTMLInputElement
    const file = input.files?.[0]

    if (!file || !uploadTarget) {
      input.value = ''
      return
    }

    try {
      await uploadCardImage(uploadTarget.index, uploadTarget.field, file)
    } catch (error) {
      uploadError = error instanceof Error ? error.message : 'Не удалось загрузить изображение.'
    } finally {
      input.value = ''
      uploadTarget = null
    }
  }

  async function handlePasteImage(event: ClipboardEvent, index: number, field: 'questionImages' | 'answerImages') {
    const file = Array.from(event.clipboardData?.items ?? [])
      .find((item) => item.type.startsWith('image/'))
      ?.getAsFile()

    if (!file) {
      return
    }

    event.preventDefault()

    try {
      await uploadCardImage(index, field, file)
    } catch (error) {
      uploadError = error instanceof Error ? error.message : 'Не удалось загрузить изображение.'
    }
  }

  function ensureCardImages(cards: CardData[]) {
    return cards.map((card) => ({
      ...card,
      questionImages: card.questionImages.map((image) => resolveImageById(image.id) ?? image),
      answerImages: card.answerImages.map((image) => resolveImageById(image.id) ?? image),
    }))
  }

  async function handleCreateSet() {
    const cards = previewMode === 'list' ? previewCards : ensureCardImages(parseCardData(sourceText))
    await onCreateSet(cards)
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
        <pre class="max-h-5 overflow-auto whitespace-pre-wrap text-sm text-muted-foreground">{promptText}</pre>
      </div>

      <div class="grid gap-4 lg:grid-cols-2 lg:items-stretch">
        <input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={handleFileInputChange} />
        <div class="flex min-h-[18rem] flex-col gap-4">
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
            <div class="flex items-center gap-2">
              {#if previewMode === 'list'}
                <Button variant="outline" size="sm" onclick={addPreviewCard} aria-label="Добавить карточку">
                  <Plus class="size-4" />
                </Button>
              {/if}
              <Button
                variant="outline"
                size="sm"
                onclick={() => setPreviewMode(previewMode === 'text' ? 'list' : 'text')}
              >
                {previewMode === 'text' ? 'Показать список' : 'Показать текст'}
              </Button>
            </div>
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
                    <section data-testid={`preview-card-${index}`} class="relative space-y-2 rounded-xl border border-border/70 bg-background/70 p-3">
                      <Button
                        variant={pendingDeleteIndex === index ? 'destructive' : 'outline'}
                        size="icon-sm"
                        class="absolute top-2 right-3"
                        onclick={() => requestDeleteCard(index)}
                        aria-label={pendingDeleteIndex === index ? 'Подтвердить удаление карточки' : 'Удалить карточку'}
                      >
                        <Trash2 class="size-4" />
                      </Button>
                      <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Карточка {index + 1}</p>
                      <label class="flex flex-col gap-1">
                        <span class="flex items-center justify-between gap-3 font-medium">
                          <span>Вопрос</span>
                          <Button variant="outline" size="sm" onclick={() => openImagePicker(index, 'questionImages')} disabled={card.questionImages.length >= 5} aria-label="Добавить изображение">
                            <ImagePlus class="size-4" />
                          </Button>
                        </span>
                        <textarea
                          value={card.question}
                          class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-15 overflow-auto resize-y rounded-xl border bg-transparent px-3 py-2 outline-none focus-visible:ring-3"
                          oninput={(event) => updateCardField(index, 'question', event.currentTarget.value)}
                          onpaste={(event) => void handlePasteImage(event, index, 'questionImages')}
                        ></textarea>
                        {#if card.questionImages.length > 0}
                          <div class="grid gap-2 sm:grid-cols-2">
                            {#each card.questionImages as image (image.id)}
                              <div class="relative overflow-hidden rounded-xl border bg-background/80 p-2">
                                <img src={`data:${image.mimeType};base64,${image.dataBase64}`} alt="Изображение вопроса" class="h-28 w-full rounded-lg object-cover" />
                                <Button variant="outline" size="icon-xs" class="absolute top-3 right-3" onclick={() => removeCardImage(index, 'questionImages', image.id)} aria-label="Удалить изображение вопроса">
                                  <X class="size-3" />
                                </Button>
                              </div>
                            {/each}
                          </div>
                        {/if}
                      </label>
                      <label class="flex flex-col gap-1">
                        <span class="flex items-center justify-between gap-3 font-medium">
                          <span>Ответ</span>
                          <Button variant="outline" size="sm" onclick={() => openImagePicker(index, 'answerImages')} disabled={card.answerImages.length >= 5} aria-label="Добавить изображение">
                            <ImagePlus class="size-4" />
                          </Button>
                        </span>
                        <textarea
                          value={card.answer}
                          class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 h-15 overflow-auto resize-y rounded-xl border bg-transparent px-3 py-2 outline-none focus-visible:ring-3"
                          oninput={(event) => updateCardField(index, 'answer', event.currentTarget.value)}
                          onpaste={(event) => void handlePasteImage(event, index, 'answerImages')}
                        ></textarea>
                        {#if card.answerImages.length > 0}
                          <div class="grid gap-2 sm:grid-cols-2">
                            {#each card.answerImages as image (image.id)}
                              <div class="relative overflow-hidden rounded-xl border bg-background/80 p-2">
                                <img src={`data:${image.mimeType};base64,${image.dataBase64}`} alt="Изображение ответа" class="h-28 w-full rounded-lg object-cover" />
                                <Button variant="outline" size="icon-xs" class="absolute top-3 right-3" onclick={() => removeCardImage(index, 'answerImages', image.id)} aria-label="Удалить изображение ответа">
                                  <X class="size-3" />
                                </Button>
                              </div>
                            {/each}
                          </div>
                        {/if}
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
      {#if createStatus}
        <p class={`text-sm ${createError ? 'text-destructive' : 'text-muted-foreground'}`}>{createStatus}</p>
      {/if}
      {#if loadLinkError}
        <p class="text-sm text-destructive">{loadLinkError}</p>
      {/if}
      {#if uploadError}
        <p class="text-sm text-destructive">{uploadError}</p>
      {/if}
    </Card.Content>
    <Card.Footer class="justify-between gap-3 max-sm:flex-col max-sm:items-stretch">
      <p class="text-sm text-muted-foreground">
        {copyState === 'done' ? 'Промпт скопирован.' : ''}
      </p>
      <div class="flex gap-3 max-sm:flex-col">
        <Button variant="outline" size="lg" onclick={onLoadLink}>Открыть по ссылке</Button>
        <Button size="lg" onclick={() => void handleCreateSet()} disabled={isCreating || !sourceText.trim()}>
          {isCreating ? createStatus || 'Создание...' : 'Создать набор'}
        </Button>
      </div>
    </Card.Footer>
  </Card.Root>
</section>
