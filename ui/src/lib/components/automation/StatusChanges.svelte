<svelte:options
	customElement={{
		tag: 'status-changes-button',
		shadow: 'none',
		props: {
			statusTrigger: { reflect: true, type: 'Object' },
            appId: {type: 'String'}
		}
	}}
/>

<script lang="ts">
	import { Accordion, AccordionItem, Button, Input, Label, P, TabItem } from "flowbite-svelte";
	import CodeEditor from "../CodeEditor.svelte";
	import { AddStatusChangeTrigger, DeleteStatusChangeTrigger, UpdateStatusChangeTrigger } from "$lib/service";
    let {statusTrigger, appId} = $props()
    let isNew: boolean = $state(true);

    const addStatusAction = () => {
		statusTrigger.actions = [
			...(statusTrigger.actions || []),
			{
				name: '',
				script: ''
			}
		];
	};

    const saveChanges = () => {
		if (isNew) {
			AddStatusChangeTrigger(appId, statusTrigger).then((value) => {
				statusTrigger = value;
				isNew = false;
			});
		} else {
			UpdateStatusChangeTrigger(appId, statusTrigger);
		}
	};

    const deleteAction = (idx: number) => {
		// ugh 3 lines to remove an item from an array...
		var tmp = $state.snapshot(statusTrigger.actions);
		tmp?.splice(idx, 1);
		statusTrigger.actions = tmp;
	};

    const deleteTrigger = () => {
		DeleteStatusChangeTrigger(appId).then(() => {
			statusTrigger = {
				actions: [
					{
						name: '',
						script: ''
					}
				]
			};
			isNew = true;
		});
	};

</script>

<TabItem open title="Status actions">
	{#if statusTrigger.actions != null && statusTrigger.actions.length > 0}
		<Accordion>
			{#each statusTrigger.actions as a, idx}
				<AccordionItem>
					{#snippet header()}{statusTrigger.actions[idx].name}{/snippet}
					<P>
						<Label for="act-name-{idx}">Name</Label>
						<Input id="act-name-{idx}" type="text" bind:value={statusTrigger.actions[idx].name}
						></Input>
					</P>
					<P>
						<Label>Script</Label>
						<CodeEditor bind:value={statusTrigger.actions[idx].script} language="risor"
						></CodeEditor>
					</P>
					<P>
						<Button onclick={() => deleteAction(idx)}>Delete action</Button>
					</P>
				</AccordionItem>
			{/each}
		</Accordion>
	{:else}
		No actions defined.
	{/if}
	<P>
		<Button color="secondary" onclick={addStatusAction}>Add action</Button>
        <Button color="secondary" onclick={saveChanges}>Save changes</Button>
		<Button onclick={deleteTrigger} hidden={isNew}>Delete trigger</Button>
	</P>
</TabItem>
