<script lang="ts">
  import { onMount } from 'svelte'
  import ArrowRight from '@lucide/svelte/icons/arrow-right'
  import Activity from '@lucide/svelte/icons/activity'
  import MoonStar from '@lucide/svelte/icons/moon-star'
  import Rocket from '@lucide/svelte/icons/rocket'
  import ShieldCheck from '@lucide/svelte/icons/shield-check'
  import Sparkles from '@lucide/svelte/icons/sparkles'

  import { API_URL } from '$lib/config'
  import * as Avatar from '$lib/components/ui/avatar'
  import { Badge } from '$lib/components/ui/badge'
  import { Button } from '$lib/components/ui/button'
  import * as Card from '$lib/components/ui/card'
  import { Input } from '$lib/components/ui/input'
  import { Separator } from '$lib/components/ui/separator'
  import { getHealth } from '$lib/api/health'

  let healthStatus = 'Loading...'
  let healthDetails = `GET ${API_URL}/api/health`

  onMount(async () => {
    try {
      const health = await getHealth()
      healthStatus = typeof health === 'string' ? health : JSON.stringify(health)
    } catch (error) {
      healthStatus = 'Request failed'
      healthDetails = error instanceof Error ? error.message : 'Unknown error'
    }
  })
</script>

<main class="min-h-screen bg-background text-foreground">
  <div class="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-8 px-4 py-8 sm:px-6 lg:px-8">
    <section class="overflow-hidden rounded-3xl border bg-card/80 shadow-sm backdrop-blur">
      <div class="grid gap-8 p-6 sm:p-8 lg:grid-cols-[1.4fr_0.9fr] lg:p-10">
        <div class="space-y-6">
          <Badge variant="secondary" class="gap-1.5">
            <MoonStar class="size-3.5" />
            Dark mode enabled
          </Badge>
          <div class="space-y-4">
            <h1 class="max-w-2xl text-4xl font-semibold tracking-tight text-balance sm:text-5xl">
              Shadcn UI starter for a clean Svelte SPA.
            </h1>
            <p class="max-w-2xl text-base text-muted-foreground sm:text-lg">
              Vite, Svelte and shadcn-inspired primitives are initialized with the default
              neutral Nova look, ready for feature work instead of boilerplate.
            </p>
          </div>
          <div class="flex flex-col gap-3 sm:flex-row">
            <Button size="lg" class="gap-2">
              Launch build
              <ArrowRight class="size-4" />
            </Button>
            <Button size="lg" variant="outline">Browse components</Button>
          </div>
        </div>

        <Card.Root class="border-border/70 bg-background/60">
          <Card.Header>
            <Card.Title>Quick subscribe</Card.Title>
            <Card.Description>Basic shadcn form controls are already wired in.</Card.Description>
          </Card.Header>
          <Card.Content class="space-y-3">
            <Input type="email" placeholder="you@example.com" />
            <Button class="w-full">Join waitlist</Button>
          </Card.Content>
        </Card.Root>
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-3">
      <Card.Root>
        <Card.Header>
          <Card.Action>
            <Rocket class="size-4 text-muted-foreground" />
          </Card.Action>
          <Card.Title>Fast start</Card.Title>
          <Card.Description>CLI-initialized Vite app with starter UI primitives added.</Card.Description>
        </Card.Header>
      </Card.Root>

      <Card.Root>
        <Card.Header>
          <Card.Action>
            <Sparkles class="size-4 text-muted-foreground" />
          </Card.Action>
          <Card.Title>Default look</Card.Title>
          <Card.Description>Neutral Nova tokens, CSS variables and Tailwind v4.</Card.Description>
        </Card.Header>
      </Card.Root>

      <Card.Root>
        <Card.Header>
          <Card.Action>
            <ShieldCheck class="size-4 text-muted-foreground" />
          </Card.Action>
          <Card.Title>Solid base</Card.Title>
          <Card.Description>Alias support, dark defaults and a verified production build.</Card.Description>
        </Card.Header>
      </Card.Root>

      <Card.Root>
        <Card.Header>
          <Card.Action>
            <Activity class="size-4 text-muted-foreground" />
          </Card.Action>
          <Card.Title>API ready</Card.Title>
          <Card.Description>Axios is configured against `API_URL` with a localhost fallback.</Card.Description>
        </Card.Header>
        <Card.Content class="space-y-2">
          <p class="text-sm text-muted-foreground">{healthDetails}</p>
          <Badge variant="outline" class="max-w-full truncate">{healthStatus}</Badge>
        </Card.Content>
      </Card.Root>
    </section>

    <Card.Root>
      <Card.Header>
        <Card.Title>Included starter components</Card.Title>
        <Card.Description>
          A few common primitives are already present for the first real screen.
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-6">
        <div class="flex flex-wrap items-center gap-3">
          <Badge>Button</Badge>
          <Badge variant="secondary">Card</Badge>
          <Badge variant="outline">Input</Badge>
          <Badge variant="ghost">Avatar</Badge>
          <Badge variant="secondary">Separator</Badge>
        </div>
        <Separator />
        <div class="flex items-center gap-3">
          <Avatar.Root size="lg">
            <Avatar.Fallback>UI</Avatar.Fallback>
          </Avatar.Root>
          <div>
            <p class="font-medium">Ready for app-specific work</p>
            <p class="text-sm text-muted-foreground">
              Replace this screen with your routes, data flows and product UI.
            </p>
          </div>
        </div>
      </Card.Content>
      <Card.Footer class="justify-between gap-3 text-sm text-muted-foreground max-sm:flex-col max-sm:items-start">
        <span>`pnpm run dev` starts the local SPA.</span>
        <span>`pnpm run build` checks the production bundle.</span>
      </Card.Footer>
    </Card.Root>
  </div>
</main>
