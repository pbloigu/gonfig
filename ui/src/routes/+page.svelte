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

		Badge


	} from 'flowbite-svelte';

	import EditConfig from '../EditConfig.svelte';
	import AddApplication from '../AddApplication.svelte';
	import DeleteApplication from '../DeleteApplication.svelte';
	import { ListApplications, Logout } from '../service';
	import ViewMeasurements from '../ViewMeasurements.svelte';

	let { data }: PageProps = $props();

	let tableItems = $state(data.apps);

	let dataChanged = () => {
		console.log('Data changed.');
		ListApplications().then((value) => {
			tableItems = value;
		});
	};

	let logout = () => {
		Logout();
	};
</script>

<Navbar>
	<NavBrand>
		<span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white"
			>Welcome to Gonfig</span
		>
	</NavBrand>
	<NavUl class="order-1">
		<NavLi>
      <div class="flex items-center space-x-1 md:order-2">
        <AddApplication {dataChanged}></AddApplication>
        <Button color="red" onclick={logout}>Logout</Button>
      </div>
    </NavLi>
	</NavUl>
</Navbar>
<Table id="hello">
	<TableHead>
		<TableHeadCell>Status</TableHeadCell>
		<TableHeadCell>Id</TableHeadCell>
		<TableHeadCell>Name</TableHeadCell>
		<TableHeadCell>Measurements</TableHeadCell>
		<TableHeadCell>Configuration</TableHeadCell>
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
				<TableBodyCell><ViewMeasurements app={ti}></ViewMeasurements></TableBodyCell>
				<TableBodyCell><EditConfig app={ti}></EditConfig></TableBodyCell>
				<TableBodyCell><DeleteApplication app={ti} {dataChanged}></DeleteApplication></TableBodyCell>
			</TableBodyRow>
		{/each}
	</TableBody>
</Table>
