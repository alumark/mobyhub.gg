<script>
  let username, password

  import { push } from 'svelte-spa-router';
  import { jwtToken } from '../stores/auth.js';
  import axios from 'axios';

  let errors = {username:'', password:'', password2: '', key:'', email: ''}
  const login = () => {
      axios
        .post("/api/users/login", { username, password })
        .then(res => {
          const { token } = res.data;
          jwtToken.set(token);
          push('/dashboard');
        })
        .catch(err => errors = err.response.data);
  }
</script>

<div>
  <div class="w-full max-w-xs object-center">
    <form class="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4" id="data">
      <div class="mb-3">
        <label
          class="block text-gray-700 text-sm font-bold mb-2"
          for="username"
        >
          Username
        </label>
        <input
          class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          id="username"
          type="text"
          placeholder="Username"
          bind:value={username}
        />
        {#if errors.username}
          <span class="text-red-700">{errors.username}</span>
        {/if}
      </div>
      <div class="mb-2">
        <label
          class="block text-gray-700 text-sm font-bold mb-2"
          for="password"
        >
          Password
        </label>
        <input
          class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          id="password"
          bind:value={password}
          type="password"
          placeholder="Password"
        />
        {#if errors.password}
          <span class="text-red-700">{errors.password}</span>
        {/if}
      </div>
      <div class="flex items-center justify-between mb-2">
        <button
          class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
          type="button"
          on:click|preventDefault={login}
        >
          Login
        </button>
        <a
          class="inline-block align-baseline font-bold text-sm text-blue-500 hover:text-blue-800"
          href='/forgot'
        >
          Forgot password?
        </a>
      </div>
    </form>
  </div>
</div>
