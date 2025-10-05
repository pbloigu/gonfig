<script lang="ts">
	import { Button, Card, Heading, Label, Navbar, NavBrand, P, Select, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import { ListSeriesValues } from '$lib/service';
	import type { SeriesValues } from '$lib/client/definitions';
	import AddSeries from '$lib/components/AddSeries.svelte';
	import { invalidateAll } from '$app/navigation';
	

	let { data }: PageProps = $props();
	let selectedSeries: string | undefined = $state();
	let tableData: SeriesValues = $state({
		page:1,
		total:0,
		pageSize:0,
		series: {
			name: "",
		},
		values: []
	})

	$effect(() => {
		if (selectedSeries) {
			ListSeriesValues(selectedSeries, data.appId, 1).then((val: SeriesValues) => {
				tableData = val
			});
		}
	});
	export const seriesAdded = function(){
		invalidateAll()
	}
	
</script>

<Navbar class="bg-primary-50 dark:bg-secondary-300">
	<NavBrand>
		<Heading tag="h6">{data.app.id}</Heading>
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
		<AddSeries app={data.app} seriesAdded={seriesAdded}></AddSeries>
	</div>
</Navbar>
<Table>
	<TableHead>
		<TableHeadCell>Time</TableHeadCell>
		<TableHeadCell>Value</TableHeadCell>
	</TableHead>
	<TableBody class="divide-y">
		{#each tableData.values as v}
			<TableBodyRow>
				<TableBodyCell>{v.time}</TableBodyCell>
				<TableBodyCell>{v.data}</TableBodyCell>
			</TableBodyRow>
		{/each}
	</TableBody>
</Table>

