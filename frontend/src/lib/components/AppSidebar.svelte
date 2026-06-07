<script lang="ts">
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import Plus from '@lucide/svelte/icons/plus'
  import Link from '@lucide/svelte/icons/link'
  import Check from '@lucide/svelte/icons/check'
  import FileText from '@lucide/svelte/icons/file-text'
  import Play from '@lucide/svelte/icons/play'

  type CardSetItem = {
    id: string
    title: string
    author: string
  }

  type RecentSessionItem = {
    id: string
    cardsetId: string
    cardsetTitle: string
    updatedAt: string
  }

  let {
    cardSets = [] as CardSetItem[],
    recentSessions = [] as RecentSessionItem[],
    activeCardsetId = '',
    activeSessionId = '',
    copyLinkState = 'idle' as 'idle' | 'done',
    onCreateCardSet = () => {},
    onCopyLink = () => {},
    onNavigate = () => {},
    formatDate = () => '',
  } = $props()
</script>

<Sidebar.Root>
  <Sidebar.Header>
    <div class="flex items-center justify-between">
      <span class="text-sm font-semibold truncate">Flashcards</span>
      <div class="flex gap-0.5">
        <Sidebar.MenuButton size="sm" onclick={onCreateCardSet} tooltipContent="New card set">
          <Plus />
        </Sidebar.MenuButton>
        <Sidebar.MenuButton size="sm" onclick={onCopyLink} tooltipContent="Copy link">
          {#if copyLinkState === 'done'}
            <Check />
          {:else}
            <Link />
          {/if}
        </Sidebar.MenuButton>
      </div>
    </div>
  </Sidebar.Header>
  <Sidebar.Separator />
  <Sidebar.Content>
    <Sidebar.Group>
      <Sidebar.GroupLabel>Card sets</Sidebar.GroupLabel>
      <Sidebar.GroupContent>
        <Sidebar.Menu>
          {#each cardSets as set (set.id)}
            <Sidebar.MenuItem>
              <Sidebar.MenuButton
                isActive={activeCardsetId === set.id}
                onclick={() => onNavigate(`/${set.id}`)}
              >
                <FileText />
                <span>{set.title}</span>
              </Sidebar.MenuButton>
            </Sidebar.MenuItem>
          {:else}
            <div class="px-3 py-2 text-xs text-muted-foreground">No card sets yet</div>
          {/each}
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>
    <Sidebar.Group>
      <Sidebar.GroupLabel>Sessions</Sidebar.GroupLabel>
      <Sidebar.GroupContent>
        <Sidebar.Menu>
          {#each recentSessions as session (session.id)}
            <Sidebar.MenuItem>
              <Sidebar.MenuButton
                isActive={activeSessionId === session.id}
                onclick={() => onNavigate(`/${session.cardsetId}/${session.id}`)}
              >
                <Play />
                <span>{session.cardsetTitle}</span>
                <span class="ml-auto text-xs text-muted-foreground">{formatDate(session.updatedAt)}</span>
              </Sidebar.MenuButton>
            </Sidebar.MenuItem>
          {:else}
            <div class="px-3 py-2 text-xs text-muted-foreground">No sessions yet</div>
          {/each}
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>
  </Sidebar.Content>
</Sidebar.Root>
