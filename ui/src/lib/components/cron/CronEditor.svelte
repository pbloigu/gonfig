<svelte:options
	customElement={{
		tag: 'cron-editor',
		shadow: 'none'
	}}
/>

<script lang="ts">
	import type { CronValidationRequest } from '$lib/client/definitions';
	import { IsValid } from '$lib/service';
	import { Helper, Input, Label } from 'flowbite-svelte';
	let expr: CronValidationRequest = $state({
		dayOfMonth: '*',
		dayOfWeek: '*',
		hour: '*',
		minute: '*',
		month: '*'
	});
	let isValid: boolean = $state(true);
	const validate = function () {
		IsValid($state.snapshot(expr)).then((valid: boolean) => {
			isValid = valid
		})
	};
</script>

<div class="flex-colflex">
	<div class="flex flex-row items-center justify-center gap-4">
		<div>
			<Label for="minute-input">Minute:</Label>
			<Input
				type="text"
				id="minute-input"
				placeholder="*"
				defaultValue="*"
				onInput={() => validate()}
				required
				bind:value={expr.minute}
			/>
		</div>
		<div>
			<Label for="hour-input">Hour:</Label>
			<Input
				type="text"
				id="hour-input"
				placeholder="*"
				defaultValue="*"
				onInput={() => validate()}
				required
				bind:value={expr.hour}
			/>
		</div>
		<div>
			<Label for="dom-input">Day of month:</Label>
			<Input
				type="text"
				id="dom-input"
				placeholder="*"
				defaultValue="*"
				onInput={() => validate()}
				required
				bind:value={expr.dayOfMonth}
			/>
		</div>
		<div>
			<Label for="month-input">Month:</Label>
			<Input
				type="text"
				id="month-input"
				placeholder="*"
				defaultValue="*"
				onInput={() => validate()}
				required
				bind:value={expr.month}
			/>
		</div>
		<div>
			<Label for="dow-input">Day of week:</Label>
			<Input
				type="text"
				id="dow-input"
				placeholder="*"
				defaultValue="*"
				onInput={() => validate()}
				required
				bind:value={expr.dayOfWeek}
			/>
		</div>
	</div>
	{#if !isValid}
	<div class="items-center justify-center flex">
		<Helper class="mt-2 text-sm" color="red">
			<span class="font-medium">Error!</span>
			Cron expression is not valid.
		</Helper>
	</div>
	{/if}
</div>
