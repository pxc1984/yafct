# frontend-svelte-shadcn

Svelte 5 + TypeScript + Vite starter using `pnpm`, Tailwind CSS v4, and real `shadcn-svelte` components.

## Requirements

- Node.js 20+
- `pnpm`
- Android Studio or a full JDK 21 installation for APK builds

## Getting Started

Install dependencies:

```bash
pnpm install
```

Optional: configure the backend API URL:

```bash
cp .env.example .env
```

Default API hosts:

- browser dev/build: `http://localhost:8080`
- Android emulator build: `http://10.0.2.2:8080`

Start the dev server:

```bash
pnpm run dev
```

Open the local URL printed by Vite.

## Available Scripts

Run the development server:

```bash
pnpm run dev
```

Build for production:

```bash
pnpm run build
```

Build and sync the Android app wrapper:

```bash
pnpm run mobile:build
```

Build a debug APK:

```bash
export JAVA_HOME=/path/to/jdk-21
pnpm run mobile:apk
```

Build a signed release APK:

```bash
export JAVA_HOME=/path/to/jdk-21
pnpm run mobile:release
```

Android App Links for `https://fc.iamamaev.ru` are enabled in the manifest.
To let Android open those links directly in the app without a chooser, publish this file on your domain:

`https://fc.iamamaev.ru/.well-known/assetlinks.json`

```json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.yafct.app",
      "sha256_cert_fingerprints": [
        "69:25:56:E0:38:0E:20:F3:EB:A4:5F:0B:85:3D:75:F0:04:EA:82:31:73:A1:FB:DB:52:C2:53:C2:78:F0:15:90"
      ]
    }
  }
]
```

Open the generated Android project in Android Studio:

```bash
pnpm run mobile:open
```

Preview the production build locally:

```bash
pnpm run preview
```

Run Svelte and TypeScript checks:

```bash
pnpm run check
```

Run ESLint:

```bash
pnpm run lint
```

## Project Structure

```text
src/
  App.svelte                 Main starter screen
  app.css                    Global Tailwind/theme styles
  main.ts                    App bootstrap, enables dark mode by default
  lib/
    api/
      client.ts              Shared Axios instance
      health.ts              Example health endpoint call
    config.ts                Frontend config values
    utils.ts                 `cn` helper and shared types
    components/ui/           Generated shadcn-svelte UI components
  vite-env.d.ts              Typed Vite env variables
components.json              shadcn-svelte CLI configuration
vite.config.ts               Vite config with Tailwind and $lib alias
vitest.config.ts             Vitest config for jsdom and test setup
tsconfig*.json               TypeScript config
eslint.config.js             ESLint config
.env.example                 Example environment variables
```

## Styling

- Tailwind CSS v4 is imported in `src/app.css`
- Geist Variable font is loaded globally
- Theme tokens are defined with CSS variables in `src/app.css`
- Dark mode is enabled by default in `src/main.ts`

## Using shadcn-svelte

This project is configured for the `shadcn-svelte` CLI through `components.json`.

Add a new component:

```bash
pnpm dlx shadcn-svelte@latest add <component-name>
```

Example:

```bash
pnpm dlx shadcn-svelte@latest add dialog
```

Overwrite an existing generated component:

```bash
pnpm dlx shadcn-svelte@latest add <component-name> -o
```

## Import Conventions

Import generated UI components from their package-style entrypoints:

```ts
import { Button } from '$lib/components/ui/button'
import * as Card from '$lib/components/ui/card'
import * as Avatar from '$lib/components/ui/avatar'
```

Use `$lib` for shared code under `src/lib`.

## API Setup

This template includes Axios and a small API layer.

Environment variable:

```bash
API_URL=http://localhost:8080
```

Behavior:

- uses `API_URL` when set
- falls back to `http://localhost:8080` when `API_URL` is empty or unset

Axios client:

```ts
import { api } from '$lib/api/client'
```

Example endpoint helper:

```ts
import { getHealth } from '$lib/api/health'

const health = await getHealth()
```

That helper calls:

```text
GET {API_URL}/api/health
```

The starter screen already shows a live example request using this setup.

## Typical Workflow

1. Install dependencies with `pnpm install`.
2. Start development with `pnpm run dev`.
3. Edit `src/App.svelte` or add routes/components for your app.
4. Add more `shadcn-svelte` components with the CLI as needed.
5. Run `pnpm run check` and `pnpm run lint` before finishing changes.
6. Run `pnpm run build` to verify the production bundle.

## Notes

- This template is a client-side Svelte SPA built with Vite.
- The initial screen is starter content and is meant to be replaced.
- If you change aliases, update both `tsconfig.json` and `vite.config.ts`.
- For Android builds, set `API_URL` to a backend reachable from the device or emulator. `localhost` inside the Android WebView points to the device itself, not your development machine.
- The default Android fallback uses `http://10.0.2.2:8080`, which maps the Android emulator back to your machine.
- For a physical Android device, set `API_URL` in `.env` to a LAN URL like `http://192.168.1.10:8080` before running `pnpm run mobile:build`.
