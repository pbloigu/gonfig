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
		Modal,
		Tabs,
		TabItem,
	} from 'flowbite-svelte';
	import StatusChanges from './StatusChanges.svelte';
	import {
		GetStatusChangeTrigger,
	} from '../../service';
	import type { Action, StatusChangeTrigger } from '$lib/client/definitions';

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

	const showDialog = () => {
		GetStatusChangeTrigger(app.id).then((value) => {
			modalStatus = true;
			if (value != null) {
				statusTrigger = value;
			}
		});
	};
</script>

<Button color="secondary" onclick={showDialog}>Automation</Button>
<Modal title="Automation" bind:open={modalStatus}>
	<Tabs>
		<StatusChanges {statusTrigger} appId={app.id}></StatusChanges>
		<TabItem title="Cron actions">No Cron actions</TabItem>
		<TabItem title="Series actions">No series actions</TabItem>
	</Tabs>
</Modal>
