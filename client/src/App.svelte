<script>
    import Router from 'svelte-spa-router';

    import Signup from './lib/Signup.svelte';
    import Home from './lib/Home.svelte';
    import Login from './lib/Login.svelte';
    import Dashboard from './lib/Dashboard.svelte';
    import NotFoundPage from './lib/NotFoundPage.svelte';

    import { jwtToken } from './stores/auth.js';

    const logout = () => {
        jwtToken.set(null);
    }
</script>

<nav class='container flex items-center fixed w-screen z-10 m-2 h-14 justify-between'>
    <div class='flex items-center rounded-md bg-white w-full h-full px-4 py-2'>
        <a href='/#/' class="home-button">
            mobyhub
        </a>
        {#if $jwtToken == 'null' || $jwtToken == null}
            <a href='/#/signup' class='nav-item'>
                Sign Up
            </a>
            
            <a href='/#/login' class='nav-item'>
                Login
            </a>
        {:else}
            <a href='/#/dashboard' class='nav-item'>
                Dashboard
            </a>
            <button on:click|preventDefault={logout} class='nav-item'>
                Logout
            </button>
        {/if}  
    </div>
</nav>
<main class='flex items-center justify-center bg-gray-300 h-screen'>
    <div class="flex items-center justify-center">
        <div class="w-full max-w-xs object-center">
            {#if $jwtToken == 'null' || $jwtToken == null}
                <Router routes={{'/login': Login, '/signup': Signup, '/': Home, '*': NotFoundPage}}>
                    
                </Router>
            {:else}
                <Router routes={{'/dashboard': Dashboard, '/': Home, '*': NotFoundPage}}>
                        
                </Router>
            {/if}
            
            <p class="text-center text-gray-500 text-xs">
                &copy;2022 Alumark. All rights reserved.
            </p>
        </div>
    </div> 
</main>

