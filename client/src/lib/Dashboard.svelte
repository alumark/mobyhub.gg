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
  <button on:click={copy}>Copy Script</button>
</Clipboard>
