<svelte:options
	customElement={{
		tag: 'add-series-button',
		shadow: 'none'
	}}
/>

<script lang="ts">
	import { AddSeries } from '$lib/service';
	import { Button, Input, Label, Modal, P } from 'flowbite-svelte';
	import type { Series } from '$lib/client/definitions';
	let { app, seriesAdded } = $props();
	let modalStatus = $state(false);
	let seriesName: string = $state('');

	const addSeries = () => {
		if (seriesName) {
			AddSeries(seriesName, app.id).then((s: Series) => {
				modalStatus = false;
				seriesAdded();
			});
		}
	};
</script>

<Button
	size="xs"
	color="secondary"
	onclick={() => {
		modalStatus = true;
	}}>Add series</Button
>
<Modal title="Add new time series" bind:open={modalStatus}>
	<form>
		<P>
			<Label for="app-name">Name</Label>
			<!-- svelte-ignore binding_property_non_reactive -->
			<Input id="app-name" type="text" bind:value={seriesName}></Input>
		</P>
	</form>
	<P class="action-buttons">
		<Button disabled={!seriesName} color="secondary" onclick={addSeries}>Save</Button>
	</P>
</Modal>
