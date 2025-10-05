<script lang="ts">
	import StatusChanges from "$lib/components/automation/StatusChanges.svelte";
	import type { PageProps } from "./$types";
  
  let { data }: PageProps = $props();
</script>

<StatusChanges appId={data.app.id}></StatusChanges>