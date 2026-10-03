<script lang="ts">
 import * as Select from '$lib/components/ui/select/index.js';
 import { Label } from '$lib/components/ui/label/index.js';
 let {id,label,value,options,disabled=false,onChange}:{id:string;label:string;value:string;options:{value:string;label:string}[];disabled?:boolean;onChange:(value:string)=>void}=$props();
</script>
<Label for={id}>{label}</Label>
<Select.Root type="single" {value} {disabled} onValueChange={onChange}>
 <Select.Trigger {id} class="min-h-11 w-full data-[size=default]:h-auto whitespace-normal text-left"><span class="min-w-0 break-words">{options.find(option=>option.value===value)?.label??''}</span></Select.Trigger>
 <Select.Content>{#each options as option}<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>{/each}</Select.Content>
</Select.Root>
