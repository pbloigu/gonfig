<script lang="ts">
	import { page } from '$app/state';
	import type { Error } from '$lib/service';
	import { P } from 'flowbite-svelte';

    let error: Error = page.state
    console.log(page.state)
</script>

<P>
    {error.code}
</P>
<P>
    {error.message}
</P>