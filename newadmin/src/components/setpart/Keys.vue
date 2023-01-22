<script setup>
</script>

<template>
  <div class="flex-1">
    <div class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center">
      <h1 class="mr-2" tabindex="-1">Keys</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>View and manage your Auth and API keys.</p>
      <p class="mt-2">Your private device keys are not included here: they are always private, stay on your device, and
        are never shared with Tailscale. <a href="https://tailscale.com/blog/tailscale-key-management/" target="_blank"
          rel="noopener noreferrer" class="link"
          aria-label="Read documentation about managing Auth and API keys">Learn&nbsp;more&nbsp;→</a></p>
    </div>
    <div class="mt-10">
      <div class="sm:flex sm:justify-between sm:items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">Auth keys</h3>
          <p class="text-gray-600">Authenticate devices without an interactive login. <a
              href="https://tailscale.com/kb/1085/auth-keys/" target="_blank" rel="noopener noreferrer" class="link"
              aria-label="Read documentation about auth keys">Learn&nbsp;more&nbsp;→</a></p>
        </div><button class="button button-outline button-small mt-2 sm:mt-0" type="button" aria-haspopup="dialog"
          aria-expanded="false" aria-controls="radix-:r1k:" data-state="closed">Generate auth key…</button>
      </div>
      <div class="rounded-md border border-gray-200 mt-4 bg-gray-0 p-6">
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">You don’t have any keys yet</div>
        </div>
      </div>
      <div class="sm:flex sm:justify-between sm:items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">API keys</h3>
          <p class="text-gray-600">API keys give access to the Tailscale API.</p>
        </div><button class="button button-outline button-small mt-2 sm:mt-0" type="button" aria-haspopup="dialog"
          aria-expanded="false" aria-controls="radix-:r1n:" data-state="closed">Generate API key…</button>
      </div>
      <div class="rounded-md border border-gray-200 mt-4 bg-gray-0 p-6">
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">You don’t have any keys yet</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>

</style>