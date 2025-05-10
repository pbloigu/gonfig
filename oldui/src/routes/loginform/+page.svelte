<script lang="ts">
    	import {
		Button,
		Input,
		Label,
		P,
        Heading
	} from 'svelte-5-ui-lib';
    import {Login} from '../../service'
    let username: string = '';
    let password: string = '';

    function login(): void {
        Login(username, password)
    }
</script>
<div class="mx-auto max-w-sm items-center gap-x-4 rounded-xl bg-white p-6 shadow-lg outline outline-black/5 dark:bg-slate-800 dark:shadow-none dark:-outline-offset-1 dark:outline-white/10">
<div>
    <Heading tag="h1">Login</Heading>
</div>
<div>
    <Label for="app-name">Username</Label>
    <Input id="app-name" type="text" bind:value={username}></Input>
</div>
<div>
    <Label for="app-name">Password</Label>
    <Input id="app-name" type="password" bind:value={password}></Input>
</div>
<div>
    <Button color="secondary" onclick={login}>Login</Button>
</div>
</div>

