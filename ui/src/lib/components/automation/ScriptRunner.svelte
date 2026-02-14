<svelte:options
	customElement={{
		tag: 'script-runner-button',
		shadow: 'none',
		props: {
			script: { type: 'String' }
		}
	}}
/>

<script lang="ts">
	import { Execute, type ScriptParam } from '$lib/service';
	import { Button, Input, Label, Modal } from 'flowbite-svelte';
	import { CircleMinusSolid, CirclePlusSolid } from 'flowbite-svelte-icons';

	let { script } = $props();
	let open: boolean = $state(false);
	let params: ScriptParam[] = $state([]);
	let response: string | null = $state(null);

	$effect(() => {
		if (!open) {
			response = null;
		} else {
			params = [
				{
					key: '',
					value: ''
				}
			];
		}
	});
	let add = function () {
		params.push({
			key: '',
			value: ''
		});
	};
	let remove = function (idx: number) {
		params.splice(idx, 1);
		if(params.length == 0) {
			params = [
				{
					key: '',
					value: ''
				}
			];
		}
	};
	let execute = function () {
		Execute(script, params).then((r) => {
			if (r == null) {
				response = 'Success!';
			} else {
				response = r;
			}
		});
	};
</script>

<Button color="secondary" onclick={() => (open = true)}>Test</Button>

<Modal bind:open size="xs">
	{#each params as param, idx}
		<fieldset class="flex gap-4">
			<Label>
				Name
				<Input required bind:value={param.key} />
			</Label>
			<Label>
				Value
				<Input required bind:value={param.value} />
			</Label>
			<div
				class="pt-8"
				tabindex="0"
				color="secondary"
				onclick={() => add()}
				style="cursor: pointer;"
				role="button"
				onkeydown={() => add()}
			>
				<CirclePlusSolid />
			</div>
			<div
				class="pt-8"
				tabindex="0"
				color="secondary"
				onclick={() => remove(idx)}
				style="cursor: pointer;"
				role="button"
				onkeydown={() => remove(idx)}
			>
				<CircleMinusSolid />
			</div>
		</fieldset>
	{/each}
	<fieldset class="border p-4">
		<legend class="px-2">Output</legend>
		{response}
	</fieldset>
	{#snippet footer()}
		<Button type="submit" color="secondary" onclick={() => execute()}>Execute</Button>
	{/snippet}
</Modal>
