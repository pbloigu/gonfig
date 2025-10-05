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
	import { AddSeries, ListSeries, ListSeriesValues } from '$lib/service';
	import type { SeriesValues } from '$lib/client/definitions';
	import { ArrowLeftOutline, ArrowRightOutline } from 'flowbite-svelte-icons';

	interface SeriesName {
		value: string;
		name: string;
	}

	let { app } = $props();

	const stateView = uiHelpers();
	let modalStatus = $state(false);
	let action: string = $state('INITIAL');
	let seriesNames: SeriesName[] = $state([]);
	let seriesValues: SeriesValues | undefined = $state();
	let newSeries: string;
	let selectedSeries: string | undefined = $state();
	let helper = $state({ start: 1, end: 10, total: 100 });
	let page = $state(1);

	$effect(() => {
		if (selectedSeries != undefined && selectedSeries.length > 0) {
			ListSeriesValues(selectedSeries, app.id, page).then((val: SeriesValues) => {
				seriesValues = val;
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

	async function getSeries() {
		seriesNames = [] as SeriesName[];
		(await ListSeries(app.id)).forEach((n) => {
			seriesNames.push({
				name: n.name,
				value: n.name
			});
		});
	}

	const showState = async () => {
		await getSeries();
		modalStatus = true;
	};
	function addSeries() {
		action = 'ADD_SERIES';
	}
	async function storeNew() {
		await AddSeries(newSeries, app.id);
		await getSeries();
		action = 'INITIAL';
	}
</script>

<Button color="secondary" onclick={showState}>View</Button>
<Modal title="App state" bind:open={modalStatus}>
	<P>App ID: {app.id}</P>
	{#if selectedSeries != undefined && selectedSeries.length > 0}
		<P>Series: {selectedSeries}</P>
	{/if}
	{#if action == 'INITIAL'}
		<P
			>Select series: <Select
				items={seriesNames}
				bind:value={selectedSeries}
				placeholder="Select series"
				class="!rounded-s-none"
			/>
			<Button size="sm" color="secondary" onclick={addSeries}>New</Button>
		</P>
	{:else if action == 'ADD_SERIES'}
		<Label for="series-name">Name</Label>
		<!-- svelte-ignore binding_property_non_reactive -->
		<Input id="series-name" type="text" bind:value={newSeries}></Input>
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
				{#if seriesValues != undefined}
					{#each seriesValues.values as v}
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
