<svelte:options
	customElement={{
		tag: 'edit-config-button',
		shadow: 'none',
		props: {
			app: { reflect: true, type: 'Object' }
		}
	}}
/>

<script lang="ts">
	import {
		Button,
		P,
		Modal,
		Textarea,
		Toolbar,
		ToolbarButton,
		ToolbarGroup,
		uiHelpers
	} from 'flowbite-svelte';
	import { CodeOutline } from 'flowbite-svelte-icons';
	import { UpdateConfiguration } from './service';
	let { app } = $props();

	
	let modalStatus = $state(false);
	
	

	const showConfig = () => {
		modalStatus = true
	};

	const saveConfig = () => {
		UpdateConfiguration(app.id, app.configuration.data).then((value) => {
			modalStatus = false
		})
	}
</script>

<Button color="secondary" onclick={showConfig}>Edit</Button>
<Modal title="App config" bind:open={modalStatus}>
	<form>
		<P>
			<!-- svelte-ignore binding_property_non_reactive -->
			<Textarea id="editor" rows={8} bind:value={app.configuration.data}>
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
</Modal>
