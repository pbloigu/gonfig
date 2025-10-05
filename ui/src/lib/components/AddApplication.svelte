<svelte:options
	customElement={{
		tag: 'add-application-button',
		shadow: 'none'
	}}
/>

<script lang="ts">
	import {
		Button,
		uiHelpers,
		Modal,
		Input,
		Label,
		P,
		Textarea,
		Toolbar,
		ToolbarGroup,
		ToolbarButton
	} from 'flowbite-svelte';
	import { CodeOutline } from 'flowbite-svelte-icons';
	import { type Application } from '$lib/client/definitions';
	import { AddApplication } from '../service';
	let { dataChanged } = $props();
	let app: Application = $state({ name: '', configuration: { data: '' } });
	

	let modalStatus = $state(false);

	
	const openDialog = () => {
		app = { name: '', configuration: { data: '' } };
		modalStatus = true
	};
	const saveApplication = () => {
		AddApplication(app).then((value) => {
			app = value;
			dataChanged();
		});
	};
</script>

<Button size="xs" color="secondary" onclick={openDialog}>Add application</Button>

<Modal title="Application details" bind:open={modalStatus}>
	{#if app.id == undefined}
		<form>
			<P>
				<Label for="app-name">Name</Label>
				<!-- svelte-ignore binding_property_non_reactive -->
				<Input id="app-name" type="text" bind:value={app.name}></Input>
			</P>
			<P>
				<Label for="editor">Configuration</Label>
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
				<Button color="secondary" onclick={saveApplication}>Save</Button>
			</P>
		</form>
	{:else}
		<P>
			<Label for="appId">Application ID</Label>
			<Input value={app.id} disabled={true}></Input>
		</P>
		<P>
			<Label for="appId">API Key</Label>
			<Input value={app.apiKey} disabled={true}></Input>
		</P>
	{/if}
</Modal>
