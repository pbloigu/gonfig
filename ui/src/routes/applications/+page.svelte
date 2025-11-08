<script lang="ts">
	import type { PageProps } from './$types';
	import {
		Table,
		TableHead,
		TableHeadCell,
		TableBody,
		TableBodyRow,
		TableBodyCell,
		Navbar,
		NavBrand,
		NavUl,
		Button,
		NavLi,
		Badge,
		A,
		Heading
	} from 'flowbite-svelte';

	
	import DeleteApplication from '$lib/components/DeleteApplication.svelte';
	import { ListApplications } from '$lib/service';
	import AddApplication from '$lib/components/AddApplication.svelte';

	let { data }: PageProps = $props();

	let tableItems = $state(data.apps);

	let dataChanged = () => {
		console.log('Data changed.');
		ListApplications().then((value) => {
			tableItems = value;
		});
	};
</script>

<Navbar class="bg-primary-50 dark:bg-secondary-300">
	<NavBrand>
		<Heading tag="h6">Applications</Heading>
	</NavBrand>

	<div class="flex md:order-2">
		<AddApplication {dataChanged}></AddApplication>
	</div>
</Navbar>

<Table>
	<TableHead>
		<TableHeadCell>Status</TableHeadCell>
		<TableHeadCell>Id</TableHeadCell>
		<TableHeadCell>Name</TableHeadCell>
		<TableHeadCell>Time series</TableHeadCell>
		<TableHeadCell>Configuration</TableHeadCell>
		<TableHeadCell>Automation</TableHeadCell>
		<TableHeadCell>Delete</TableHeadCell>
	</TableHead>
	<TableBody>
		{#each tableItems as ti}
			<TableBodyRow>
				<TableBodyCell>
					{#if ti.isOnline}
						<Badge color="green">ONLINE</Badge>
					{:else}
						<Badge color="red">OFFLINE</Badge>
					{/if}
				</TableBodyCell>
				<TableBodyCell>{ti.id}</TableBodyCell>
				<TableBodyCell>{ti.name}</TableBodyCell>				
				<TableBodyCell><A href="/applications/{ti.id}/series">Time series</A></TableBodyCell>
				<TableBodyCell><A href="/applications/{ti.id}/configuration">Configuration</A></TableBodyCell>				
				 <TableBodyCell><A href="/applications/{ti.id}/automation">Automation</A></TableBodyCell>
				<TableBodyCell><DeleteApplication app={ti} {dataChanged}></DeleteApplication></TableBodyCell>
			</TableBodyRow>
		{/each}
	</TableBody>
</Table>
