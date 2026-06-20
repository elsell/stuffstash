<script lang="ts">
  export let onSignIn: () => Promise<void>;
</script>

<section class="signin-panel">
  <div>
    <h2>Sign in with local Dex</h2>
    <p>
      The web tracer bullet uses the same OIDC boundary as the API. Start the local Dex stack, then sign in as a
      local test user.
    </p>
  </div>
  <button type="button" onclick={onSignIn}>Sign in</button>
</section>
