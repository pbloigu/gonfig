<script lang="ts">
	import '../app.css';
	let { children } = $props();
	import { Darkmode } from 'svelte-5-ui-lib';
</script>
<Darkmode />
{@render children()}
