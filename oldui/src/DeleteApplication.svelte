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
	import { Button, P, Modal, uiHelpers } from 'svelte-5-ui-lib';
	import { DeleteApplication } from './service';
	let { app, dataChanged } = $props();

	const confirmView = uiHelpers();

	let modalStatus = $state(false);
	const closeModal = confirmView.close;
	$effect(() => {
		modalStatus = confirmView.isOpen;
	});

	const showConfirm = () => {
		confirmView.open();
	};
	const doDelete = () => {
		DeleteApplication(app.id).then((value) => {
			dataChanged();
			closeModal();
		});
	};
</script>

<Button color="red" onclick={showConfirm}>Delete</Button>
<Modal title="Confirm application delete" {modalStatus} {closeModal}>
	<P>
		Application {app.id} will be deleted along with any configuration and state it has.
	</P>
	<P>Are you sure?</P>
	<P>
		<Button color="red" onclick={doDelete}>Yes</Button><Button
			color="alternative"
			onclick={closeModal}>No</Button
		>
	</P>
</Modal>
