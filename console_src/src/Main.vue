<script setup>
</script>
<template>
      <div class="relative flex min-h-screen flex-col justify-center bg-gray-200">
    <div
      class="absolute inset-0 bg-[url(/img/grid.svg)] bg-center [mask-image:linear-gradient(180deg,white,rgba(255,255,255,0))]">
    </div>
    <div class="md:w-2/3 self-center card lg:card-side bg-base-100 shadow-xl">
      <figure><img src="/img/logoname.png" alt="Logo" /></figure>
      <div class="card-body mx-auto lg:flex-shrink-0 w-96">
        <h1 class="card-title">蜃境</h1>
        <span class="badge badge-secondary">仅供测试</span>
        <div class="divider"></div>

        <div class="card-actions justify-end pt-4">
          <!--
          <a href="/admin">
            <button class="btn btn-primary">注册</button>
          </a>
          -->
          <a href="/admin">
            <button class="btn btn-primary">进入控制台</button>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>