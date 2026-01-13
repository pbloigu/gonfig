<script lang="ts">
	import { DarkMode } from "flowbite-svelte";
    import '../app.css';
	import Navigation from "$lib/components/Navigation.svelte";
	import { token } from "$lib/service";
	
	let { children } = $props();
</script>

<DarkMode />
<div class="min-w-full">
    {#if token.current.value}
        <Navigation />
    {/if}
	{@render children()}
</div>
