<script setup>
</script>

<template>
  <div class="flex-1">
    <div class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center">
      <h1 class="mr-2" tabindex="-1">密钥管理</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>查看和管理您的<strong>授权密钥</strong>和<strong>API密钥</strong></p>
      <p class="mt-2">您个人设备的私钥<strong>不</strong>在此管理，它们永远是<strong>私有</strong>并存在于您的设备上，<strong>不</strong>向蜃境披露</p>
    </div>
    <div class="mt-10">
      <div class="sm:flex sm:justify-between sm:items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">授权密钥</h3>
          <p class="text-gray-600">用于无需交互式登录即可实现设备鉴别接入 </p>
        </div>
        <button class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3">
        生成授权密钥…</button>
      </div>
      <div class="rounded-md border border-gray-200 mt-4 bg-gray-0 p-6">
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">你还没有任何密钥</div>
        </div>
      </div>
      <div class="sm:flex sm:justify-between sm:items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">API 密钥</h3>
          <p class="text-gray-600">API 密钥用于访问蜃境API.</p>
        </div><button class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3">
        生成 API 密钥…</button>
      </div>
      <div class="rounded-md border border-gray-200 mt-4 bg-gray-0 p-6">
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">你还没有任何密钥</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>

</style>