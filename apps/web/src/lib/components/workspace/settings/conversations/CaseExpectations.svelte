<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import type { CaseDefinition, CaseExpectations, CaseOperation, CaseProposal } from '$lib/domain/conversationCase';
  import { fixtureParentChoices } from '$lib/application/caseFixtureEditing';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Label from '$lib/components/ui/label/index.js';
  import * as Textarea from '$lib/components/ui/textarea/index.js';
  import WorkflowSelect from './WorkflowSelect.svelte';
  import ValidationMessage from './ValidationMessage.svelte';
  import { validationAttributes } from './validationPresentation';
  let { value, errors = {}, disabled = false, onChange }: { value: CaseDefinition; errors?: Record<string, string>; disabled?: boolean; onChange: (value: CaseDefinition) => void } = $props();
  const operations: { value: CaseOperation; label: string }[] = [
    { value: 'create', label: t('web.CaseExpectations.create') }, { value: 'move', label: t('web.CaseExpectations.move') }, { value: 'archive', label: t('web.CaseExpectations.archive') },
    { value: 'restore', label: t('web.CaseExpectations.restore') }, { value: 'checkout', label: t('web.CaseExpectations.checkOut') }, { value: 'return', label: t('web.CaseExpectations.return') }
  ];
  const fixtures = $derived(value.assets.map(asset => ({ value: asset.id, label: asset.title || t('web.CaseExpectations.unnamedFixture') })));
  function change(patch: Partial<CaseExpectations>) { if (!disabled) onChange({ ...value, expectations: { ...value.expectations, ...patch } }); }
  function reference(id: string) { const current = value.expectations.referencedAssets; change({ referencedAssets: current.includes(id) ? current.filter(value => value !== id) : [...current, id] }); }
  function forbidden(operation: CaseOperation) { const current = value.expectations.forbiddenOperations; change({ forbiddenOperations: current.includes(operation) ? current.filter(value => value !== operation) : [...current, operation] }); }
  function proposal(index: number, patch: Partial<CaseProposal>) { change({ proposals: value.expectations.proposals.map((item, current) => current === index ? { ...item, ...patch } : item) }); }
  function operation(index: number, name: string) {
    const valid = operations.find(candidate => candidate.value === name); if (!valid) return;
    proposal(index, { operation: valid.value, targetId: '', destinationId: '', newTitle: '', newKind: valid.value === 'create' ? 'item' : '', details: '' });
  }
  function ancestors(id: string) {
    const byId = new Map(value.assets.map(asset => [asset.id, asset])); const seen = new Set([id]);
    const choices: { value: string; label: string }[] = []; let parent = byId.get(id)?.parentId;
    while (parent && !seen.has(parent)) { seen.add(parent); const asset = byId.get(parent); if (!asset) break;
      choices.push({ value: asset.id, label: asset.title || t('web.CaseExpectations.unnamedFixture') }); parent = asset.parentId; }
    return choices;
  }
</script>
<section class="case-expectations" aria-labelledby="case-expectations-title">
  <h3 id="case-expectations-title" tabindex="-1" {...validationAttributes(errors['case-expectations-title'], 'case-expectations-title')}>{t('web.CaseExpectations.expectedResult')}</h3><ValidationMessage field="case-expectations-title" message={errors['case-expectations-title']} />
  <WorkflowSelect id="expected-outcome" error={errors['expected-outcome']} label={t('web.CaseExpectations.theConversationShould')} value={value.expectations.kind} {disabled}
    options={[{ value: 'answer', label: t("web.options.CaseExpectations.answerTheQuestion") }, { value: 'clarification', label: t("web.options.CaseExpectations.askForClarification") }, { value: 'proposal', label: t("web.options.CaseExpectations.proposeAChange") }, { value: 'failure', label: t("web.options.CaseExpectations.reportAFailure") }]}
    onChange={kind => { if (kind === 'answer' || kind === 'clarification' || kind === 'proposal' || kind === 'failure') change({ kind }); }} />
  <fieldset {disabled}><legend>{t('web.CaseExpectations.mustReferToTheseFixtures')}</legend><div class="choices">
    {#each value.assets as asset (asset.id)}<Button.Root type="button" variant="outline" aria-pressed={value.expectations.referencedAssets.includes(asset.id)} onclick={() => reference(asset.id)}>{asset.title || t('web.CaseExpectations.unnamedFixture')}</Button.Root>{/each}
  </div>{#if !value.assets.length}<p>{t('web.CaseExpectations.addFixturesToRequireSpecificItemsInTheAnswer')}</p>{/if}</fieldset>
  <fieldset {disabled}><legend>{t('web.CaseExpectations.expectedLocations')}</legend>
    {#each value.expectations.locations as location, index}<div class="expectation-row">
      <WorkflowSelect id={`location-item-${index}`} error={errors[`location-item-${index}`]} label={t('web.CaseExpectations.item')} value={location.assetId} {disabled} options={fixtures}
        onChange={assetId => change({ locations: value.expectations.locations.map((item, current) => current === index ? { assetId, ancestorId: '' } : item) })} />
      <WorkflowSelect id={`location-parent-${index}`} error={errors[`location-parent-${index}`]} label={t('web.CaseExpectations.mustBeInside')} value={location.ancestorId} {disabled} options={ancestors(location.assetId)}
        onChange={ancestorId => change({ locations: value.expectations.locations.map((item, current) => current === index ? { ...item, ancestorId } : item) })} />
      <Button.Root type="button" variant="outline" onclick={() => change({ locations: value.expectations.locations.filter((_, current) => current !== index) })}>{t('web.CaseExpectations.removeLocationExpectation')}</Button.Root>
    </div>{/each}
    <Button.Root type="button" variant="outline" disabled={disabled || !value.assets.length || value.expectations.locations.length >= 100} onclick={() => change({ locations: [...value.expectations.locations, { assetId: '', ancestorId: '' }] })}>{t('web.CaseExpectations.addExpectedLocation')}</Button.Root>
  </fieldset>
  {#if value.expectations.kind === 'proposal' || value.expectations.proposals.length}
    <fieldset {disabled}><legend>{t('web.CaseExpectations.expectedProposedChanges')}</legend>
      {#if value.expectations.kind !== 'proposal'}<p>{t('web.CaseExpectations.removeTheseChangesOrChooseProposeAChangeBefore')}</p>{/if}
      {#each value.expectations.proposals as expected, index}<div class="expectation-row">
        <WorkflowSelect id={`proposal-operation-${index}`} error={errors[`proposal-operation-${index}`]} label={t('web.CaseExpectations.change')} value={expected.operation} {disabled} options={operations} onChange={name => operation(index, name)} />
        {#if expected.operation === 'create'}
          <Label.Root class="grid gap-2">{t('web.CaseExpectations.newItemName')}<Input.Root name={`proposal-title-${index}`} id={`proposal-title-${index}`} {...validationAttributes(errors[`proposal-title-${index}`], `proposal-title-${index}`)} value={expected.newTitle} required oninput={event => proposal(index, { newTitle: event.currentTarget.value })} /></Label.Root><ValidationMessage field={`proposal-title-${index}`} message={errors[`proposal-title-${index}`]} />
          <WorkflowSelect id={`proposal-kind-${index}`} error={errors[`proposal-kind-${index}`]} label={t('web.CaseExpectations.newItemKind')} value={expected.newKind} {disabled} options={[{ value: 'item', label: t("web.options.CaseExpectations.item") }, { value: 'container', label: t("web.options.CaseExpectations.container") }, { value: 'location', label: t("web.options.CaseExpectations.location") }]}
            onChange={newKind => { if (newKind === 'item' || newKind === 'container' || newKind === 'location') proposal(index, { newKind }); }} />
        {:else}<WorkflowSelect id={`proposal-target-${index}`} error={errors[`proposal-target-${index}`]} label={t('web.CaseExpectations.existingItem')} value={expected.targetId} {disabled} options={fixtures} onChange={targetId => proposal(index, { targetId, destinationId: '' })} />{/if}
        {#if expected.operation === 'create' || expected.operation === 'move'}
          <WorkflowSelect id={`proposal-destination-${index}`} error={errors[`proposal-destination-${index}`]} label={t('web.CaseExpectations.destination')} value={expected.destinationId} {disabled}
            options={[{ value: '', label: t("web.options.CaseExpectations.noParent") }, ...fixtureParentChoices(value.assets, expected.operation === 'move' ? expected.targetId : '').map(asset => ({ value: asset.id, label: asset.title || t("web.options.CaseExpectations.unnamedFixture") }))]}
            onChange={destinationId => proposal(index, { destinationId })} />
        {/if}
        {#if expected.operation === 'checkout' || expected.operation === 'return'}<Label.Root class="grid gap-2">{t('web.CaseExpectations.expectedDetails')}<Textarea.Root name={`proposal-details-${index}`} id={`proposal-details-${index}`} {...validationAttributes(errors[`proposal-details-${index}`], `proposal-details-${index}`)} value={expected.details} oninput={event => proposal(index, { details: event.currentTarget.value })} /></Label.Root><ValidationMessage field={`proposal-details-${index}`} message={errors[`proposal-details-${index}`]} />{/if}
        <Button.Root type="button" variant="outline" onclick={() => change({ proposals: value.expectations.proposals.filter((_, current) => current !== index) })}>{t('web.CaseExpectations.removeProposedChange')}</Button.Root>
      </div>{/each}
      <Button.Root type="button" variant="outline" disabled={disabled || value.expectations.proposals.length >= 100} onclick={() => change({ proposals: [...value.expectations.proposals, { operation: 'create', targetId: '', destinationId: '', newKind: 'item', newTitle: '', details: '' }] })}>{t('web.CaseExpectations.addProposedChange')}</Button.Root>
    </fieldset>
  {/if}
  <fieldset {disabled}><legend>{t('web.CaseExpectations.mustNeverProposeTheseChanges')}</legend><div class="choices">
    {#each operations as option}<Button.Root type="button" variant="outline" aria-pressed={value.expectations.forbiddenOperations.includes(option.value)} onclick={() => forbidden(option.value)}>{option.label}</Button.Root>{/each}
  </div><p>{t('web.CaseExpectations.testRunsNeverApproveOrExecuteInventoryChanges')}</p></fieldset>
</section>
<style>
  .case-expectations, fieldset, .expectation-row { display: grid; gap: 1rem; min-width: 0; }
  fieldset { border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; }
  .expectation-row + .expectation-row { border-top: 1px solid var(--border); padding-top: 1rem; }
  h3, legend { font-weight: 600; } p { color: var(--muted-foreground); font-size: .9rem; }
  .choices { display: flex; gap: .5rem; flex-wrap: wrap; }
</style>
