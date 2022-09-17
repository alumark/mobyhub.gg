<script>
    import { user } from '../stores/auth';
</script>

<div class="my-2">
    <p class="text-center m-2">
        A modern script hub
    </p>
    
    <span class="flex justify-center">
        {#if $user == 'nulll' || $user == null}
            <a href='/#/login' class='bg-blue-500 hover:bg-blue-700 mx-1 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline'>Log in</a>
            <a href='/#/signup' class='bg-blue-500 hover:bg-blue-700 mx-1 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline'>Sign up</a>
        {:else}
        <a href='/#/dashboard' class='bg-blue-500 hover:bg-blue-700 mx-1 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline'>Dashboard</a>
        {/if}
    </span>
</div>