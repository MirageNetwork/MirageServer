<script setup>
</script>

<template>
          <div class="flex-1">
        <div class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center">
          <h1 class="mr-2" tabindex="-1">通用</h1>
        </div>
        <div class="text-gray-600 mt-3 mb-10">
          <p>Manage your tailnet settings.</p>
        </div>
        <div class="mt-10">
          <div class="space-y-10">
            <div>
              <header class="max-w-2xl">
                <h3 class="text-xl font-semibold tracking-tight">Organization </h3>
                <p class="mt-1 text-gray-600">This is your organization’s name. It cannot be changed. </p>
              </header>
              <div class="mt-4">
                <div class="max-w-sm">
                  <div class="flex input-wrapper relative overflow-hidden min-w-0"><input
                      class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis" type="text"
                      readonly="" value="nopkt.com"><button
                      class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans link font-medium text-sm"><span
                        class="absolute">Copy</span><span class="invisible">Copied!</span></button></div>
                </div>
              </div>
            </div>
            <div>
              <header class="max-w-2xl">
                <h3 class="text-xl font-semibold tracking-tight">Identity Provider </h3>
                <p class="mt-1 text-gray-600">How users log in to your tailnet. <a
                    href="https://tailscale.com/contact/support" class="link" target="_blank"
                    rel="noopener noreferrer">Contact support</a> to change this. </p>
              </header>
              <div class="mt-4">
                <div class="rounded border border-gray-300 bg-gray-0 py-2 px-4 inline-flex items-center font-medium">
                  <span class="mr-2"><svg width="17" height="17" viewBox="0 0 16 16" fill="none"
                      xmlns="http://www.w3.org/2000/svg">
                      <path d="M0 0H7.57886V7.57886H0V0Z" fill="#F25022"></path>
                      <path d="M0 8.42114H7.57886V16H0V8.42114Z" fill="#00A4EF"></path>
                      <path d="M8.42114 0H16V7.57886H8.42114V0Z" fill="#7FBA00"></path>
                      <path d="M8.42114 8.42114H16V16H8.42114V8.42114Z" fill="#FFB900"></path>
                    </svg></span>Microsoft
                </div>
              </div>
            </div>
            <div>
              <header class="max-w-2xl">
                <h3 class="text-xl font-semibold tracking-tight">Device Authorization </h3>
                <p class="mt-1 text-gray-600">Require new devices to be approved by admins before they can access the
                  network. <a href="https://tailscale.com/kb/1099/device-authorization" target="_blank"
                    rel="noopener noreferrer" class="link"
                    aria-label="Read documentation about Device Authorization">Learn&nbsp;more&nbsp;→</a></p>
              </header>
              <div class="mt-4"><span data-state="closed">
                  <div class="flex items-center"><input id="require-approval" type="checkbox" class="toggle mr-3"
                      checked=""><label class="font-medium cursor-pointer" for="require-approval">Manually authorize new
                      devices</label></div>
                </span></div>
            </div>
            <div>
              <header class="max-w-2xl">
                <h3 class="text-xl font-semibold tracking-tight">Key Expiry </h3>
                <p class="mt-1 text-gray-600">Set the number of days a device can stay logged in to Tailscale before it
                  needs to re-authenticate with Microsoft. <a href="https://tailscale.com/kb/1028/key-expiry"
                    target="_blank" rel="noopener noreferrer" class="link"
                    aria-label="Read documentation about Key Expiry">Learn&nbsp;more&nbsp;→</a></p>
              </header>
              <div class="mt-4"><span data-state="closed">
                  <div class="flex">
                    <div class="relative"><input class="input z-10 rounded-r-none" type="text" inputmode="numeric"
                        pattern="[0-9]*" id="key-expiry-duration" value="180">
                      <div class="bg-white top-1 bottom-1 right-1 rounded-r-md absolute flex items-center">
                        <div class="flex items-center"><button class="button button-minimal button-small px-2"
                            type="button" tabindex="-1"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18"
                              viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                              stroke-linecap="round" stroke-linejoin="round">
                              <line x1="5" y1="12" x2="19" y2="12"></line>
                            </svg></button><button class="button button-minimal button-small px-2" disabled=""
                            type="button" tabindex="-1"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18"
                              viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                              stroke-linecap="round" stroke-linejoin="round">
                              <line x1="12" y1="5" x2="12" y2="19"></line>
                              <line x1="5" y1="12" x2="19" y2="12"></line>
                            </svg></button></div>
                      </div>
                    </div>
                    <div
                      class="flex items-center px-3 bg-gray-50 text-gray-500 border rounded-r border-l-0 border-gray-300">
                      days</div>
                  </div>
                </span>
                <p class="text-sm text-gray-500 mt-1">Must be between 1 and 180 days.</p>
                <div class="mt-4"><button class="button button-blue" disabled="" type="submit">Save</button><button
                    class="button button-outline ml-3" disabled="">Reset</button></div>
              </div>
            </div>
          </div>
          <div class="mt-8">
            <header class="max-w-2xl">
              <h3 class="text-xl font-semibold tracking-tight">Delete tailnet </h3>
              <p class="mt-1 text-gray-600">Permanently delete this tailnet. This action cannot be undone. <a
                  href="https://tailscale.com/kb/1145/remove-team-members/#deleting-your-user-account-or-your-entire-tailnet"
                  target="_blank" rel="noopener noreferrer" class="link"
                  aria-label="Read documentation about Delete tailnet">Learn&nbsp;more&nbsp;→</a></p>
            </header>
            <div class="mt-4"><span data-state="closed"><button class="button button-red button-medium">Delete
                  tailnet…</button></span></div>
          </div>
        </div>
      </div>
</template>

<style>
</style>