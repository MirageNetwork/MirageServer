<script setup></script>
<template>
  <div class="relative flex min-h-screen flex-col justify-center bg-gray-200">
    <div
      class="absolute inset-0 bg-[url(/img/grid.svg)] bg-center [mask-image:linear-gradient(180deg,white,rgba(255,255,255,0))]"
    ></div>
    <div class="md:w-2/3 self-center card lg:card-side bg-base-100 shadow-xl">
      <figure><img src="/img/logo_withname@1024.png" alt="Logo" /></figure>
      <div class="card-body mx-auto lg:flex-shrink-0 w-96">
        <div class="card-actions justify-end pt-4">
          <a href="/admin">
            <button class="btn btn-primary">进入控制台</button>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>
