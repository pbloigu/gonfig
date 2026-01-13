<svelte:options
	customElement={{
		tag: 'navigation-bar',
		shadow: 'none'
	}}
/>
<script>
	import { Button, Heading, Navbar, NavBrand } from "flowbite-svelte";
     import { goto } from '$app/navigation';
    import { Logout } from '$lib/service';
    let logout = () => {
		Logout();
	};

</script>


<Navbar class="bg-primary-200 dark:bg-secondary-700">
	<NavBrand>
		<Heading onclick={() => goto("/applications")} class="logo-text">Gonfig</Heading>
	</NavBrand>

	<div class="flex md:order-2">
		<Button color="red" onclick={logout}>Logout</Button>
	</div>
</Navbar>