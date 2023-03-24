<script setup>
import { watch, ref, onMounted, onBeforeUpdate, computed } from "vue";
import { useDisScroll } from "/src/utils.js";

useDisScroll();

onMounted(() => {});
</script>

<template>
  <div
    @click.self="$emit('close')"
    class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
    style="pointer-events: auto"
  >
    <div
      class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
      style="pointer-events: auto"
    >
      <header class="flex items-center justify-between space-x-4 mb-4 mr-8">
        <div class="font-semibold text-lg truncate">如何获取 Microsoft 认证配置信息</div>
      </header>
      <div class="mb-4">
        <h3 class="font-semibold text-gray-800 mb-2">
          第一步：在Azure Active Directory注册应用
        </h3>
        <p class="text-gray-700">首先你应该已经注册并拥有能访问Azure的账户</p>
        <p class="text-gray-700 mb-2">
          在Azure<a
            class="text-blue-400"
            href="https://portal.azure.com/#view/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/~/RegisteredApps"
            >应用注册</a
          >页面点击新注册
        </p>
        <div class="rounded-md border bg-stone-100 border-stone-300 p-2 mb-2">
          <div class="flex justify-center">
            <img class="w-full max-w-xl" src="/img/help/microsoft/createApp.jpg" />
          </div>
        </div>
        <p class="text-gray-700 mb-2">注册应用完成后的页面即显示了Client ID信息</p>
        <div class="rounded-md border bg-stone-100 border-stone-300 p-2">
          <div class="flex justify-center">
            <img class="w-full max-w-xl" src="/img/help/microsoft/getClientID.jpg" />
          </div>
        </div>
      </div>
      <div>
        <h3 class="font-semibold flex items-center text-gray-800 mb-2">
          第二步：创建Client Secret
        </h3>
        <p class="text-gray-700 mb-2">注册应用完成后页面右半部分点击进入创建客户端机密</p>
        <div class="rounded-md border bg-stone-100 border-stone-300 p-2">
          <div class="flex justify-center mb-2">
            <img
              class="w-full max-w-xl"
              src="/img/help/microsoft/createClientSecret.jpg"
            />
          </div>
          <div class="flex justify-center">
            <img class="w-full max-w-xl" src="/img/help/microsoft/getClientSecret.jpg" />
          </div>
        </div>
      </div>

      <button
        @click="$emit('close')"
        class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
        type="button"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="1.25em"
          height="1.25em"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.tooltip {
  --tooltip-color: #faf9f8;
  --tooltip-text-color: #3a3939;
  text-align: start;
  white-space: normal;
}

.tooltip:before {
  max-width: 16rem;
  font-size: small;
  font-weight: 300;
  border-radius: 0.375rem;
  box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  padding-left: 0.75rem;
  padding-right: 0.75rem;
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  border-width: 1px;
  border-color: #e1dfde;
}
</style>
