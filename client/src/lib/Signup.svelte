<script>
  import { push } from 'svelte-spa-router';
  import axios from 'axios';

  let username, password, email, key

  let errors = {username:'', password:'', key:'', email: ''}
  function signup() {
    axios
      .post("https://mobyhub.herokuapp.com/api/users/register", {username, email, password, password2: password, key})
      .then(() => push("/login")) // re-direct to login on successful register
      .catch(err =>
        errors = err.response.data
      );
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
      <div class="mb-3">
        <label
          class="block text-gray-700 text-sm font-bold mb-2"
          for="email"
        >
          Email
        </label>
        <input
          class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          id="email"
          type="text"
          placeholder="Email"
          bind:value={email}
        />
        {#if errors.email}
          <span class="text-red-700">{errors.email}</span>
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
      <div class="mb-6">
        <label
          class="block text-gray-700 text-sm font-bold mb-2"
          for="text"
        >
          Key
        </label>
        <input
          class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          id="key"
          bind:value={key}
          type="text"
          placeholder="Key"
        />
        {#if errors.key}
          <span class="text-red-700">{errors.key}</span>
        {/if}
      </div>
      <div class="flex items-center justify-between mb-2">
        <button
          class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
          type="button"
          on:click|preventDefault={signup}
        >
          Sign Up
        </button>
        <a
          class="inline-block align-baseline font-bold text-sm text-blue-500 hover:text-blue-800"
          href="https://sellix.io/product/5ee3fb669c00b"
        >
          Don't have a key?
        </a>
      </div>
    </form>
  </div>
</div>
