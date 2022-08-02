<script>
    import { user } from '../stores/auth.js';
    import Clipboard from "svelte-clipboard";
</script>

Hello, {$user} <br/>

<Clipboard
  text='loadstring(game:HttpGet("https://mobyhub.herokuapp.com/loadstring", true))()'
  let:copy
  on:copy={() => {
    console.log('Has Copied');
  }}>
  <button on:click={copy} class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline">Copy Script</button>
</Clipboard>
