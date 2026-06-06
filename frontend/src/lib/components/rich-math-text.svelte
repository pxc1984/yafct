<script lang="ts">
  import katex from 'katex'

  type Segment =
    | { type: 'text'; value: string }
    | { type: 'math'; value: string; displayMode: boolean }

  let { text = '', class: className = '' }: { text?: string; class?: string } = $props()

  function pushText(segments: Segment[], value: string) {
    if (value) {
      segments.push({ type: 'text', value })
    }
  }

  function parseMathSegments(value: string) {
    const segments: Segment[] = []
    let cursor = 0

    while (cursor < value.length) {
      const displayStart = value.indexOf('$$', cursor)
      const bracketStart = value.indexOf('\\[', cursor)
      const inlineParenStart = value.indexOf('\\(', cursor)
      const inlineDollarStart = value.indexOf('$', cursor)

      const candidates = [
        { index: displayStart, token: '$$', closing: '$$', displayMode: true },
        { index: bracketStart, token: '\\[', closing: '\\]', displayMode: true },
        { index: inlineParenStart, token: '\\(', closing: '\\)', displayMode: false },
        { index: inlineDollarStart, token: '$', closing: '$', displayMode: false },
      ].filter((candidate) => candidate.index >= 0)

      if (candidates.length === 0) {
        pushText(segments, value.slice(cursor))
        break
      }

      candidates.sort((a, b) => a.index - b.index)
      const next = candidates[0]

      if (next.token === '$' && value[next.index + 1] === '$') {
        cursor = next.index + 1
        continue
      }

      pushText(segments, value.slice(cursor, next.index))

      const contentStart = next.index + next.token.length
      const contentEnd = value.indexOf(next.closing, contentStart)

      if (contentEnd < 0) {
        pushText(segments, value.slice(next.index))
        break
      }

      segments.push({
        type: 'math',
        value: value.slice(contentStart, contentEnd).trim(),
        displayMode: next.displayMode,
      })

      cursor = contentEnd + next.closing.length
    }

    return segments
  }

  const segments = $derived(parseMathSegments(text))
</script>

<div class={`math-content whitespace-pre-wrap ${className}`.trim()}>
  {#each segments as segment, index (index)}
    {#if segment.type === 'text'}
      {segment.value}
    {:else}
      <!-- eslint-disable-next-line svelte/no-at-html-tags -->
      {@html katex.renderToString(segment.value, {
        throwOnError: false,
        displayMode: segment.displayMode,
        strict: 'ignore',
      })}
    {/if}
  {/each}
</div>
