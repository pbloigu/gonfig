<script lang="ts">
	import StatusChanges from "$lib/components/automation/StatusChanges.svelte";
	import { Heading, Navbar, NavBrand } from "flowbite-svelte";
	import type { PageProps } from "./$types";
  
  let { data }: PageProps = $props();
</script>
<Navbar class="bg-primary-50 dark:bg-secondary-300">
	<NavBrand>
		<Heading tag="h6">{data.app.name}</Heading>
	</NavBrand>
</Navbar>
<div class="gonfig-content">
	<StatusChanges appId={data.app.id}></StatusChanges>
</div>