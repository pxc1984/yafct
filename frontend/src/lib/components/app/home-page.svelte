<script lang="ts">
  import Copy from '@lucide/svelte/icons/copy'

  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'

  let {
    promptText,
    sourceText = $bindable(''),
    setTitle = $bindable(''),
    setDescription = $bindable(''),
    setAuthor = $bindable(''),
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

        <label class="flex min-h-[28rem] flex-col gap-2">
          <span class="text-sm font-medium">Текст с карточками</span>
          <textarea
            bind:value={sourceText}
            class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 min-h-0 flex-1 resize-none overflow-auto rounded-2xl border bg-transparent px-4 py-3 text-sm outline-none focus-visible:ring-3"
            placeholder={sourcePlaceholder}
          ></textarea>
        </label>
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
