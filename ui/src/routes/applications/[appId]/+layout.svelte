<script lang="ts">
	

	let { children, data } = $props();
	console.log(data);
</script>

<div>
	
	{@render children()}
</div>
