<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import {
    assetTagDisplayNameByteLength,
    assetTagDisplayNameMaxLength,
    assetTagKeyFromDisplayName,
    type AssetTag,
    type AssetTagDraft
  } from '$lib/domain/inventory';
  import AssetTagChips from './AssetTagChips.svelte';

  let {
    tags = [],
    selectedIds = [],
    newTags = [],
    onSelectedIdsChange,
    onNewTagsChange
  }: {
    tags?: AssetTag[];
    selectedIds?: string[];
    newTags?: AssetTagDraft[];
    onSelectedIdsChange: (ids: string[]) => void;
    onNewTagsChange: (tags: AssetTagDraft[]) => void;
  } = $props();

  let newTagName = $state('');
  let newTagColor = $state('');
  let selected = $derived(new Set(selectedIds));
  let selectedExistingTags = $derived(tags.filter((tag) => selected.has(tag.id)));
  let hasSelection = $derived(selectedExistingTags.length > 0 || newTags.length > 0);
  let normalizedNewTagColor = $derived(normalizeColor(newTagColor));
  let newTagKey = $derived(assetTagKeyFromDisplayName(newTagName));
  let newTagDisplayNameByteLength = $derived(assetTagDisplayNameByteLength(newTagName));
  let matchingExistingTag = $derived(tags.find((tag) => tag.key === newTagKey));
  let matchingPendingTag = $derived(newTags.find((tag) => assetTagKeyFromDisplayName(tag.displayName) === newTagKey));
  let canAddTag = $derived(
    newTagKey.length > 0 &&
      (matchingExistingTag !== undefined ||
        matchingPendingTag !== undefined ||
        (newTagDisplayNameByteLength <= assetTagDisplayNameMaxLength &&
          (newTagColor.trim().length === 0 || normalizedNewTagColor !== undefined)))
  );

  $effect(() => {
    const reconciledIds: string[] = [];
    const remainingTags: AssetTagDraft[] = [];
    for (const tag of newTags) {
      const tagKey = assetTagKeyFromDisplayName(tag.displayName);
      const existing = tags.find((candidate) => candidate.key === tagKey);
      if (existing) {
        reconciledIds.push(existing.id);
      } else {
        remainingTags.push(tag);
      }
    }
    if (reconciledIds.length === 0) {
      return;
    }
    onSelectedIdsChange(Array.from(new Set([...selectedIds, ...reconciledIds])));
    onNewTagsChange(remainingTags);
  });

  function toggleTag(tagId: string): void {
    const next = new Set(selectedIds);
    if (next.has(tagId)) {
      next.delete(tagId);
    } else {
      next.add(tagId);
    }
    onSelectedIdsChange(Array.from(next));
  }

  function addTag(): void {
    const displayName = newTagName.trim();
    if (!displayName) {
      return;
    }
    if (matchingExistingTag) {
      onSelectedIdsChange(Array.from(new Set([...selectedIds, matchingExistingTag.id])));
      newTagName = '';
      newTagColor = '';
      return;
    }
    if (matchingPendingTag) {
      newTagName = '';
      newTagColor = '';
      return;
    }
    if (assetTagDisplayNameByteLength(displayName) > assetTagDisplayNameMaxLength) {
      return;
    }
    if (newTagColor.trim().length > 0 && !normalizedNewTagColor) {
      return;
    }
    const color = normalizedNewTagColor;
    onNewTagsChange([...newTags, color ? { displayName, color } : { displayName }]);
    newTagName = '';
    newTagColor = '';
  }

  function removeNewTag(index: number): void {
    onNewTagsChange(newTags.filter((_, currentIndex) => currentIndex !== index));
  }

  function normalizeColor(value: string): string | undefined {
    const raw = value.trim();
    if (!raw) {
      return undefined;
    }
    const color = raw.startsWith('#') ? raw : `#${raw}`;
    return /^#[0-9a-fA-F]{6}$/.test(color) ? color.toUpperCase() : undefined;
  }

</script>

<fieldset class="tag-selector">
  <legend>Tags</legend>
  {#if hasSelection}
    <div class="selected-tag-summary">
      <AssetTagChips tags={selectedExistingTags} compact />
      {#each newTags as tag, index}
        <span class="tag-chip pending-tag">
          {#if tag.color}
            <span class="tag-swatch" style={`--tag-color: ${tag.color}`} aria-hidden="true"></span>
          {/if}
          <span>{tag.displayName}</span>
          <Button.Root type="button" variant="ghost" size="icon-sm" aria-label={`Remove ${tag.displayName}`} onclick={() => removeNewTag(index)}>
            <X />
          </Button.Root>
        </span>
      {/each}
    </div>
  {/if}

  {#if tags.length > 0}
    <div class="tag-options" aria-label="Available tags">
      {#each tags as tag}
        <Button.Root
          type="button"
          variant={selected.has(tag.id) ? 'default' : 'outline'}
          class="tag-option"
          aria-pressed={selected.has(tag.id)}
          onclick={() => toggleTag(tag.id)}
        >
          {#if tag.color}
            <span class="tag-swatch" style={`--tag-color: ${tag.color}`} aria-hidden="true"></span>
          {/if}
          {tag.displayName}
        </Button.Root>
      {/each}
    </div>
  {/if}

  <div class="new-tag-row">
    <div class="field-stack">
      <Label for="new-tag-name">New tag</Label>
      <Input id="new-tag-name" bind:value={newTagName} placeholder="Workshop" />
    </div>
    <div class="field-stack color-field">
      <Label for="new-tag-color">Color</Label>
      <Input id="new-tag-color" bind:value={newTagColor} placeholder="#2F80ED" />
    </div>
    <Button.Root type="button" variant="outline" disabled={!canAddTag} onclick={addTag}><Plus /> Add</Button.Root>
  </div>
</fieldset>
