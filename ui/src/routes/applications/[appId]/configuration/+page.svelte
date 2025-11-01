<script lang="ts">
	import { UpdateConfiguration } from '$lib/service';
	import {
		Button,

		Heading,

		Navbar,

		NavBrand,

		P,

		Textarea,
		Toolbar,
		ToolbarButton,
		ToolbarGroup
	} from 'flowbite-svelte';
	import { CodeOutline } from 'flowbite-svelte-icons';
	import type { PageProps } from './$types';

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
<Navbar class="bg-primary-50 dark:bg-secondary-300">
	<NavBrand>
		<Heading tag="h6">{data.app.name}</Heading>
	</NavBrand>
</Navbar>
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
