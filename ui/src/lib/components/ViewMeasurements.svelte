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
	import {
		Button,
		P,
		Modal,
		uiHelpers,
		Select,
		Label,
		Input,
		Table,
		TableHead,
		TableHeadCell,
		TableBody,
		TableBodyRow,
		TableBodyCell,
		Pagination
	} from 'flowbite-svelte';
	import { AddMeasurement, ListMeasurements, ListMeasurementValues } from '$lib/service';
	import type { MeasurementValues } from '$lib/client/definitions';
	import { ArrowLeftOutline, ArrowRightOutline } from 'flowbite-svelte-icons';

	interface MeasurementName {
		value: string;
		name: string;
	}

	let { app } = $props();

	const stateView = uiHelpers();
	let modalStatus = $state(false);
	let action: string = $state('INITIAL');
	let measurementNames: MeasurementName[] = $state([]);
	let measurementValues: MeasurementValues | undefined = $state();
	let newMeasurement: string;
	let selectedMeasurement: string | undefined = $state();
	let helper = $state({ start: 1, end: 10, total: 100 });
	let page = $state(1);

	$effect(() => {
		if (selectedMeasurement != undefined && selectedMeasurement.length > 0) {
			ListMeasurementValues(selectedMeasurement, app.id, page).then((val: MeasurementValues) => {
				measurementValues = val;
				helper.total = val.total;
				helper.start = (val.page - 1) * val.pageSize + 1;
				helper.end = helper.start + Math.min(val.values.length, val.pageSize) - 1;
			});
			action = 'SHOW_VALUES';
		}
	});

	const previous = async () => {
		if (page > 1) {
			page = page - 1;
		} else {
			page = 1;
		}
	};

	const next = async () => {
		page++;
	};

	async function getMeasurements() {
		measurementNames = [] as MeasurementName[];
		(await ListMeasurements(app.id)).forEach((n) => {
			measurementNames.push({
				name: n.name,
				value: n.name
			});
		});
	}

	const showState = async () => {
		await getMeasurements();
		modalStatus = true;
	};
	function addMeasurement() {
		action = 'ADD_MEASUREMENT';
	}
	async function storeNew() {
		await AddMeasurement(newMeasurement, app.id);
		await getMeasurements();
		action = 'INITIAL';
	}
</script>

<Button color="secondary" onclick={showState}>View</Button>
<Modal title="App state" bind:open={modalStatus}>
	<P>App ID: {app.id}</P>
	{#if selectedMeasurement != undefined && selectedMeasurement.length > 0}
		<P>Measurement: {selectedMeasurement}</P>
	{/if}
	{#if action == 'INITIAL'}
		<P
			>Select measurement: <Select
				items={measurementNames}
				bind:value={selectedMeasurement}
				placeholder="Select measurement"
				class="!rounded-s-none"
			/>
			<Button size="sm" color="secondary" onclick={addMeasurement}>New</Button>
		</P>
	{:else if action == 'ADD_MEASUREMENT'}
		<Label for="measurement-name">Name</Label>
		<!-- svelte-ignore binding_property_non_reactive -->
		<Input id="measurement-name" type="text" bind:value={newMeasurement}></Input>
		<Button size="sm" color="secondary" onclick={storeNew}>Add</Button>
	{:else if action == 'SHOW_VALUES'}
		<div class="flex flex-col items-center justify-center gap-3">
			<div class="flex flex-col items-center justify-center gap-2">
				<div class="text-sm text-gray-700 dark:text-gray-400">
					Showing <span class="font-semibold text-gray-900 dark:text-white">{helper.start}</span>
					to
					<span class="font-semibold text-gray-900 dark:text-white">{helper.end}</span>
					of
					<span class="font-semibold text-gray-900 dark:text-white">{helper.total}</span>
					Entries
				</div>

				<Pagination {previous} {next}>
					{#snippet prevContent()}
						<div class="flex items-center gap-2 bg-gray-800 text-white">
							<ArrowLeftOutline class="me-2 h-5 w-5" />
							Prev
						</div>
					{/snippet}
					{#snippet nextContent()}
						<div class="flex items-center gap-2 bg-gray-800 text-white">
							Next
							<ArrowRightOutline class="ms-2 h-5 w-5" />
						</div>
					{/snippet}
				</Pagination>
			</div>
		</div>
		<Table divClass="table-wrp block max-h-70 overflow-y-auto overscroll-contain" class="w-full">
			<TableHead class="sticky top-0 border-b bg-white">
				<TableHeadCell>Recored</TableHeadCell>
				<TableHeadCell>Value</TableHeadCell>
			</TableHead>
			<TableBody class="h-96 overflow-y-auto">
				{#if measurementValues != undefined}
					{#each measurementValues.values as v}
						<TableBodyRow style="height:1em">
							<TableBodyCell>{v.time}</TableBodyCell>
							<TableBodyCell>{v.data}</TableBodyCell>
						</TableBodyRow>
					{/each}
				{/if}
			</TableBody>
		</Table>
	{:else}
		ERROR
	{/if}
</Modal>
