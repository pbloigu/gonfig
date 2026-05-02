<svelte:options
	customElement={{
		tag: 'status-changes-button',
		shadow: 'none',
		props: {
			appId: { type: 'String' }
		}
	}}
/>

<script lang="ts">
	import {
		Accordion,
		AccordionItem,
		Button,
		Input,
		Label,
		P,
		Alert,
		TabItem
	} from 'flowbite-svelte';
	import CodeEditor from '../CodeEditor.svelte';
	import {
		AddStatusChangeTrigger,
		DeleteStatusChangeTrigger,
		GetStatusChangeTrigger,
		UpdateStatusChangeTrigger
	} from '$lib/service';
	import { onMount } from 'svelte';
	import ScriptRunner from './ScriptRunner.svelte';
	import type { components } from '$lib/client/api';
	let { appId } = $props();
	let isNew: boolean = $state(true);
	let openIndex: number = $state(0);
	let statusTrigger: components['schemas']['StatusChangeTrigger'] = $state({});

	onMount(() => {
		GetStatusChangeTrigger(appId).then((result) => {
			if (result) {
				statusTrigger = result;
				isNew = false;
			}
		});
	});

	const addStatusAction = () => {
		if (!statusTrigger.actions){
			statusTrigger.actions = []
		}
		statusTrigger.actions?.push({
			description: '',
			script: ''
		} as components["schemas"]["Action"]);
		openIndex = (statusTrigger.actions?.length || 1) - 1;
	};

	const saveChanges = () => {
		if (isNew) {
			AddStatusChangeTrigger(appId, statusTrigger).then((value) => {
				statusTrigger = value;
				isNew = false;
			});
		} else {
			UpdateStatusChangeTrigger(appId, statusTrigger);
		}
	};

	const deleteAction = (idx: number) => {
		// ugh 3 lines to remove an item from an array...
		var tmp = $state.snapshot(statusTrigger.actions);
		tmp?.splice(idx, 1);
		statusTrigger.actions = tmp;
	};

	const deleteTrigger = () => {
		DeleteStatusChangeTrigger(appId).then(() => {
			statusTrigger = {
				actions: [
					{
						description: '',
						script: ''
					} as components["schemas"]["Action"]
				]
			};
			isNew = true;
		});
	};
</script>

<TabItem open title="Status changes">
	{#if statusTrigger.actions != null && statusTrigger.actions.length > 0}
		<Accordion>
			{#each statusTrigger.actions as _, idx}
				{#if statusTrigger.actions[idx]}
					<AccordionItem classes={{ content: 'accordion-open' }} open={idx == openIndex}>
						{#snippet header()}{statusTrigger.actions?.[idx]?.description}{/snippet}
						<P>
							<Label for="act-name-{idx}">Name</Label>
							<Input
								id="act-name-{idx}"
								type="text"
								bind:value={statusTrigger.actions[idx].description}
							></Input>
						</P>
						<P>
							<Label for="editor-{idx}">Script</Label>
							<CodeEditor
								id="editor-{idx}"
								bind:value={statusTrigger.actions[idx].script}
								language="risor"
							></CodeEditor>
						</P>
						<P class="action-buttons">
							<Button onclick={() => deleteAction(idx)}>Delete action</Button>
							<ScriptRunner script={statusTrigger.actions[idx].script}></ScriptRunner>
						</P>
					</AccordionItem>
				{/if}
			{/each}
		</Accordion>
	{:else}
		<Alert>
			<span class="font-medium">No actions defined!</span>
			Add actions below.
		</Alert>
	{/if}
	<P class="action-buttons">
		<Button color="secondary" onclick={addStatusAction}>Add action</Button>
		<Button color="secondary" onclick={saveChanges}>Save changes</Button>
		<Button onclick={deleteTrigger} hidden={isNew}>Delete trigger</Button>
	</P>
</TabItem>
