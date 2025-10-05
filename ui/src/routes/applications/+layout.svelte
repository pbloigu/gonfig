<script lang="ts">
	import { A, Button, DarkMode, Heading, Navbar, NavBrand, NavLi, NavUl } from 'flowbite-svelte';
	import { AddApplication, Logout } from '$lib/service';
	import { goto } from '$app/navigation';
	let { children } = $props();

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
{@render children()}
