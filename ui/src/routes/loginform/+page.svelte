<script lang="ts">
	import { Button, Input, Label, P, Heading, Card } from 'flowbite-svelte';
	import { Login } from '../../lib/service';
	let username: string = '';
	let password: string = '';

	function login(): void {
		Login(username, password);
	}
</script>

<div class="w-1/3 m-auto">
	<Card class="p-4 sm:p-6 md:p-8">
		<div>
			<Heading tag="h1" class="mb-4 text-4xl font-extrabold  md:text-5xl lg:text-6xl">Login</Heading
			>
		</div>

		<form class="flex flex-col space-y-6">
			<div class="mb-6">
				<Label for="username" class="mb-2 block">Username</Label>
				<Input id="username" type="text" bind:value={username}></Input>
			</div>
			<div class="mb-6">
				<Label for="password" class="mb-2 block">Password</Label>
				<Input id="password" type="password" bind:value={password}></Input>
			</div>
			<div>
				<Button color="secondary" onclick={login}>Login</Button>
			</div>
		</form>
	</Card>
</div>
