<script lang="ts">
	import '../app.css';
	import { DarkMode } from 'flowbite-svelte';
	let { children } = $props();
</script>
<DarkMode />
<div class="min-w-full">
{@render children()}
</div>


