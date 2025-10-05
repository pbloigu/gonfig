<script lang="ts">
	import { DarkMode } from "flowbite-svelte";
    import '../app.css';
    let { children } = $props();

</script>
<DarkMode />
<div class="min-w-full">
    {@render children()}
</div>