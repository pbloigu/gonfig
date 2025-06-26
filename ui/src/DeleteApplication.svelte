<svelte:options
	customElement={{
		tag: 'delete-application-button',
		shadow: 'none',
		props: {
			app: { reflect: true, type: 'Object' }
		}
	}}
/>

<script lang="ts">
	import { Button, P, Modal, uiHelpers } from 'flowbite-svelte';
	import { DeleteApplication } from './service';
	let { app, dataChanged } = $props();

	let modalStatus = $state(false);

	const showConfirm = () => {
		modalStatus = true
	};
	const doDelete = () => {
		DeleteApplication(app.id).then((value) => {
			dataChanged();
			modalStatus = false;
		});
	};
</script>

<Button color="red" onclick={showConfirm}>Delete</Button>
<Modal title="Confirm application delete" bind:open={modalStatus}>
	<P>
		Application {app.id} will be deleted along with any configuration and state it has.
	</P>
	<P>Are you sure?</P>
	<P>
		<Button color="red" onclick={doDelete}>Yes</Button><Button
			color="alternative"
			onclick={() => {modalStatus = false}}>No</Button
		>
	</P>
</Modal>
