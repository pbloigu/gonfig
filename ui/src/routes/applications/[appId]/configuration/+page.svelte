<script lang="ts">
	import {
		Button,
	
		P,
		
		Textarea,
		Toolbar,
		ToolbarButton,
		ToolbarGroup
	} from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import { CodeOutline } from 'flowbite-svelte-icons';
	import { UpdateConfiguration } from '$lib/service';

	let { data }: PageProps = $props();
	let config: string = $state(data.app.configuration?.data || "")
	
	const saveConfig = async () => {
		if(!data.app.configuration) {
			data.app.configuration = {
				data: ""
			}
		}
		data.app.configuration.data = config
		await UpdateConfiguration(data.app.id || "", data.app.configuration?.data || "");
	};
</script>

<form>
	<P>
		<!-- svelte-ignore binding_property_non_reactive -->
		<Textarea id="editor" rows={8} bind:value={config}>
			{#snippet header()}
				<Toolbar embedded>
					<ToolbarGroup>
						<ToolbarButton name="Format code"><CodeOutline /></ToolbarButton>
					</ToolbarGroup>
				</Toolbar>
			{/snippet}
		</Textarea>
	</P>
	<P>
		<Button color="secondary" onclick={saveConfig}>Save</Button>
	</P>
</form>
