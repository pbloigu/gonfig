<svelte:options
	customElement={{
		tag: 'open-state-button',
		shadow: 'none',
		props: {
			app: { reflect: true, type: 'Object' }
		}
	}}
/>

<script lang="ts">
	import { Button, P, Modal, uiHelpers, Select } from 'svelte-5-ui-lib';
	import { ListMeasurements } from './service';

	interface MeasurementName {
		value: string
		name: string
	}

	let { app } = $props();

	const stateView = uiHelpers();
	let modalStatus = $state(false);
	let measurementNames: MeasurementName[] = $state([]);
	const closeModal = stateView.close;
	$effect(() => {
		modalStatus = stateView.isOpen;
	});

	const showState = async () => {
		(await ListMeasurements(app.id)).forEach((n) => {
			measurementNames.push({
				name: n.name,
				value: n.name
			})
		})
		
		stateView.open();
	};
</script>

<Button color="secondary" onclick={showState}>View</Button>
<Modal title="App state" {modalStatus} {closeModal}>
	<P>App ID: {app.id}</P>
	<P>Select measurement: <Select items={measurementNames} placeholder="Select measurement" class="!rounded-s-none" /></P>
</Modal>
