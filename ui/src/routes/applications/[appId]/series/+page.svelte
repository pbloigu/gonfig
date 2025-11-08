<script lang="ts">
	import { Heading, Navbar, NavBrand, Select } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import { ListSeriesValues } from '$lib/service';
	import type { SeriesValues } from '$lib/client/definitions';
	import AddSeries from '$lib/components/AddSeries.svelte';
	import { invalidateAll } from '$app/navigation';
	import { onMount } from 'svelte';

	let table: any | undefined = undefined;
	let tableElement = $state<HTMLTableElement>();
	let { data }: PageProps = $props();
	let selectedSeries: string | undefined = $state();
	const headings: string[] = ['time', 'value'];
	const options = {
		searchable: true,
		sortable: true,
		perPage: 10,
		classes: {
			active: 'datatable-active',
			bottom: 'datatable-bottom',
			container: 'datatable-container',
			cursor: 'datatable-cursor',
			dropdown: 'datatable-dropdown',
			ellipsis: 'datatable-ellipsis',
			empty: 'datatable-empty',
			headercontainer: 'datatable-headercontainer',
			info: 'datatable-info',
			input: 'datatable-input',
			loading: 'datatable-loading',
			pagination: 'datatable-pagination',
			paginationList: 'datatable-pagination-list',
			search: 'datatable-search',
			selector: 'datatable-selector',
			sorter: 'datatable-sorter',
			table: 'datatable-table',
			top: 'datatable-top',
			wrapper: 'datatable-wrapper'
		}
	};

	const createTable = async () => {
		const { DataTable } = await import('simple-datatables');
		if (tableElement) {
			table = new DataTable(tableElement, options);
		}
	};

	onMount(() => {
		createTable();
	});

	$effect(() => {
		if (selectedSeries) {
			table.destroy();
			table.init();

			ListSeriesValues(selectedSeries, data.appId, 1).then((val: SeriesValues) => {
				let rows: any[][] = [];
				val.values.forEach((v) => {
					rows.push([v.time, v.data]);
					console.log(v.data + '  ' + v.time);
				});
				table.insert({
					headings: headings,
					data: rows
				});
			});
		}
	});
	export const seriesAdded = function () {
		invalidateAll();
	};
</script>

<Navbar class="bg-primary-50 dark:bg-secondary-300">
	<NavBrand>
		<Heading tag="h6">{data.app.name}</Heading>
	</NavBrand>
	<div class="flex md:order-1">
		<Select
			id="select-series"
			items={data.series}
			bind:value={selectedSeries}
			placeholder="Select series"
			class="!rounded-s-none"
		/>
	</div>
	<div class="flex md:order-2">
		<AddSeries app={data.app} {seriesAdded}></AddSeries>
	</div>
</Navbar>
<div class="relative overflow-x-auto gonfig-content">
	<table bind:this={tableElement}></table>
</div>
