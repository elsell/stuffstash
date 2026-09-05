<script lang="ts">
  import type { CaseDefinition } from '$lib/domain/conversationCase';
  let { value }: { value: CaseDefinition } = $props();
  const title = (id: string) => value.assets.find(asset => asset.id === id)?.title || 'Unselected fixture';
</script>
<section class="case-summary">
  <h3>{value.title}</h3><p class="request">{value.utterance}</p>
  <h4>Test inventory</h4><ul>{#each value.assets as asset}<li><strong>{asset.title}</strong> · {asset.kind}{asset.parentId ? ` · inside ${title(asset.parentId)}` : ''}
    {#if asset.description}<p>{asset.description}</p>{/if}{#if asset.tagNames.length}<p>Tags: {asset.tagNames.join(', ')}</p>{/if}</li>{/each}</ul>
  <h4>Expected result: {value.expectations.kind}</h4>
  {#if value.expectations.referencedAssets.length}<p>Must mention: {value.expectations.referencedAssets.map(title).join(', ')}</p>{/if}
  <ul>{#each value.expectations.locations as location}<li>{title(location.assetId)} is inside {title(location.ancestorId)}.</li>{/each}
    {#each value.expectations.proposals as proposal}<li>{proposal.operation === 'checkout' ? 'Check out' : proposal.operation}: {proposal.operation === 'create' ? `${proposal.newTitle} (${proposal.newKind})` : title(proposal.targetId)}{proposal.destinationId ? ` → ${title(proposal.destinationId)}` : ''}{proposal.details ? ` · ${proposal.details}` : ''}</li>{/each}
  </ul>
  {#if value.expectations.forbiddenOperations.length}<p>Forbidden changes: {value.expectations.forbiddenOperations.join(', ')}</p>{/if}
</section>
<style>.case-summary { display: grid; gap: .5rem; } h3, h4 { font-weight: 600; } .request { white-space: pre-wrap; } li, p { overflow-wrap: anywhere; }</style>
