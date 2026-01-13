<script lang="ts">
	import { goto } from "$app/navigation";
	import { token } from "$lib/service";

    if(!token.current.value){
        goto("/login")
    }
</script>