<script lang="ts">
 import { Label } from '$lib/components/ui/label/index.js';
 import * as Select from '$lib/components/ui/select/index.js';
 let {label,emptyLabel,value,choices,onChange}:{label:string;emptyLabel:string;value?:string;choices:{id:string;title:string}[];onChange:(value:string|undefined)=>void}=$props();
 const id=$props.id();
</script>
<div class="choice"><Label for={id}>{label}</Label><Select.Root type="single" value={value??''} onValueChange={value=>onChange(value||undefined)}><Select.Trigger {id} class="min-h-11 w-full">{choices.find(choice=>choice.id===value)?.title ?? emptyLabel}</Select.Trigger><Select.Content><Select.Item value="" label={emptyLabel}>{emptyLabel}</Select.Item>{#each choices as choice}<Select.Item value={choice.id} label={choice.title}>{choice.title}</Select.Item>{/each}</Select.Content></Select.Root></div>
<style>.choice{display:grid;gap:.4rem;}</style>
