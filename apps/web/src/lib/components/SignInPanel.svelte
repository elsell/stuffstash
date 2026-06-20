<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';

  export let onSignIn: () => Promise<void>;
</script>

<Card.Root class="signin-panel">
  <Card.Header class="p-0">
    <Card.Title>Sign in with local Dex</Card.Title>
    <Card.Description>
      The web tracer bullet uses the same OIDC boundary as the API. Start the local Dex stack, then sign in as a
      local test user.
    </Card.Description>
  </Card.Header>
  <Button type="button" onclick={onSignIn}>Sign in</Button>
</Card.Root>
