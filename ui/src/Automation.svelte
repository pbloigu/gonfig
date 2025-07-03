<svelte:options
	customElement={{
		tag: 'edit-automation-button',
		shadow: 'none',
		props: {
			app: { reflect: true, type: 'Object' }
		}
	}}
/>

<script lang="ts">
	import {
		Button,
		P,
		Modal,
		Textarea,
		Toolbar,
		ToolbarButton,
		ToolbarGroup,
		uiHelpers,
		Tabs,
		TabItem,
		Label,
		Input
	} from 'flowbite-svelte';
	import CodeEditor from '$lib/components/code-editor.svelte';
	import { CodeOutline } from 'flowbite-svelte-icons';
	import {
		AddStatusChangeTrigger,
		DeleteStatusChangeTrigger,
		GetStatusChangeTrigger,
		UpdateConfiguration,
		UpdateStatusChangeTrigger
	} from './service';
	import type { Action, StatusChangeTrigger } from './client/definitions';

	let { app } = $props();

	let modalStatus = $state(false);
	let statusTrigger: StatusChangeTrigger = $state({
		actions: [
			{
				name: '',
				script: ''
			}
		]
	});

	let isNew: boolean = $state(true);
	let mode = localStorage.getItem('THEME_PREFERENCE_KEY');

	const showDialog = () => {
		GetStatusChangeTrigger(app.id).then((value) => {
			modalStatus = true;
			if (value != null) {
				statusTrigger = value;
				isNew = false;
			}
		});
	};

	const addStatusAction = () => {
		statusTrigger.actions = [
			...(statusTrigger.actions || []),
			{
				name: '',
				script: ''
			}
		];
	};

	const saveChanges = () => {
		if (isNew) {
			AddStatusChangeTrigger(app.id, statusTrigger).then((value) => {
				statusTrigger = value;
				isNew = false;
			});
		} else {
			UpdateStatusChangeTrigger(app.id, statusTrigger);
		}
	};

	const deleteAction = (idx: number) => {
		// ugh 3 lines to remove an item from an array...
		var tmp = $state.snapshot(statusTrigger.actions);
		tmp?.splice(idx, 1);
		statusTrigger.actions = tmp;
	};

	const deleteTrigger = () => {
		DeleteStatusChangeTrigger(app.id).then(() => {
			statusTrigger = {
				actions: [
					{
						name: '',
						script: ''
					}
				]
			};
			isNew = true;
		});
	};
</script>

<Button color="secondary" onclick={showDialog}>Automation</Button>
<Modal title="Automation" bind:open={modalStatus}>
	<Tabs>
		<TabItem open title="Status actions">
			{#if statusTrigger.actions != null && statusTrigger.actions.length > 0}
				{#each statusTrigger.actions as a, idx}
					<P>
						<Label for="act-name-{idx}">Name</Label>
						<Input id="act-name-{idx}" type="text" bind:value={statusTrigger.actions[idx].name}
						></Input>
					</P>
					<P>
						<Label>Script</Label>
						<CodeEditor
							bind:value={statusTrigger.actions[idx].script}
							language="risor"
						></CodeEditor>
					</P>
					<P>
						<Button onclick={() => deleteAction(idx)}>Delete action</Button>
					</P>
				{/each}
			{:else}
				No actions defined.
			{/if}
			<P>
				<Button color="secondary" onclick={addStatusAction}>Add action</Button>
				<Button color="secondary" onclick={saveChanges}>Save changes</Button>
				<Button onclick={deleteTrigger} hidden={isNew}>Delete trigger</Button>
			</P>
		</TabItem>
		<TabItem title="Cron actions"></TabItem>
		<TabItem title="Measurement actions"></TabItem>
	</Tabs>
</Modal>
