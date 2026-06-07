<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right'
  import Play from '@lucide/svelte/icons/play'
  import Plus from '@lucide/svelte/icons/plus'
  import Trash2 from '@lucide/svelte/icons/trash-2'

  import type { CardSet } from '$lib/api/flashcards'
  import { Badge } from '$lib/components/ui/badge'
  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'

  type SessionRecord = {
    id: string
    updatedAt: string
  }

  let {
    cardsetId,
    cardSetDetails,
    cardSetLoading,
    cardSetError,
    sessions,
    sessionListError,
    sessionsLoading,
    isMobile,
    copyLinkState,
    onCreateSession,
    onCopyLink,
    onNavigate,
    onRemoveSession,
    formatDate,
  }: {
    cardsetId: string
    cardSetDetails: CardSet | null
    cardSetLoading: boolean
    cardSetError: string
    sessions: SessionRecord[]
    sessionListError: string
    sessionsLoading: boolean
    isMobile: boolean
    copyLinkState: 'idle' | 'done'
    onCreateSession: () => void | Promise<void>
    onCopyLink: () => void | Promise<void>
    onNavigate: (path: string) => void
    onRemoveSession: (sessionId: string) => void
    formatDate: (value: string) => string
  } = $props()
</script>

<section class="mx-auto flex w-full max-w-4xl flex-1 items-center">
  <div class="grid w-full gap-6 lg:grid-cols-[0.95fr_1.05fr]">
    <Card.Root>
      <Card.Header>
        <Badge variant="outline" class="w-fit">Набор</Badge>
        <Card.Title class="break-words text-2xl">{cardSetDetails?.title || cardsetId}</Card.Title>
        {#if cardSetDetails?.author}
          <Card.Description>Автор: {cardSetDetails.author}</Card.Description>
        {/if}
      </Card.Header>
      <Card.Content class="space-y-4">
        {#if cardSetLoading}
          <div class="rounded-2xl border bg-background/60 p-4 text-sm text-muted-foreground">Загрузка описания...</div>
        {:else if cardSetError}
          <p class="text-sm text-destructive">{cardSetError}</p>
        {:else if cardSetDetails?.description}
          <div class="max-h-[7.5rem] overflow-auto rounded-2xl border bg-background/60 p-4 text-sm text-muted-foreground whitespace-pre-wrap">
            {cardSetDetails.description}
          </div>
        {/if}

        <Button size="lg" class="w-full gap-2" onclick={onCreateSession} disabled={sessionsLoading}>
          <Play class="size-4" />
          {sessionsLoading ? 'Создание...' : 'Начать новую сессию'}
        </Button>
        {#if isMobile}
          <Button variant="outline" class="w-full" onclick={onCopyLink}>
            {copyLinkState === 'done' ? 'Ссылка скопирована' : 'Копировать ссылку на набор'}
          </Button>
        {/if}
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

        {#each sessions as session (session.id)}
          <div class="flex items-center gap-2 rounded-2xl border bg-background/60 px-2 py-2 transition hover:bg-muted/60">
            <button
              class="flex min-w-0 flex-1 items-center justify-between px-2 py-1 text-left"
              onclick={() => onNavigate(`/${cardsetId}/${session.id}`)}
            >
              <div class="min-w-0">
                <p class="truncate font-medium">Сессия {session.id}</p>
                <p class="text-sm text-muted-foreground">Обновлена {formatDate(session.updatedAt)}</p>
              </div>
              <ChevronRight class="size-4 shrink-0 text-muted-foreground" />
            </button>

            <Button
              variant="ghost"
              size="icon-sm"
              aria-label="Удалить сессию"
              onclick={(event) => {
                event.stopPropagation()
                onRemoveSession(session.id)
              }}
            >
              <Trash2 class="size-4" />
            </Button>
          </div>
        {/each}
      </Card.Content>
    </Card.Root>
  </div>

  <Button
    size="icon-lg"
    class="fixed right-4 bottom-4 z-20 rounded-full shadow-lg sm:right-6 sm:bottom-6"
    onclick={() => onNavigate('/')}
    aria-label="Создать новый набор карточек"
  >
    <Plus class="size-5" />
  </Button>
</section>
