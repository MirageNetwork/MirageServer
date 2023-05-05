<script setup>
import { watch, ref, onMounted, computed } from "vue";

const props = defineProps({
  downloadDetails: Object,
});

onMounted(() => {
  console.log("windows page mounted");
  console.log("windows page got download details", props.downloadDetails);
});
</script>
<template>
  <div>
    <div class="pt-4 pb-10 border-b border-gray-200 max-w-xl mx-auto">
      <div
        v-if="
          downloadDetails.macos.primary == '' && downloadDetails.macos.secondary == ''
        "
        class="text-center"
      >
        <button
          disabled
          class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white disabled:bg-blue-600/60 disabled:text-white/60"
        >
          暂未提供macOS版本
        </button>
      </div>
      <div v-if="downloadDetails.macos.primary != ''" class="text-center">
        <a :href="downloadDetails.macos.primary">
          <button class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white">
            下载 <span class="hidden md:inline">蜃境客户端</span> macOS版
          </button>
        </a>
        <p class="mt-2 text-sm text-gray-600">要求macOS High Sierra 10.13 及以上版本</p>
      </div>
      <div
        v-if="
          downloadDetails.macos.primary == '' && downloadDetails.macos.secondary != ''
        "
        class="text-center"
      >
        <a :href="downloadDetails.macos.secondary">
          <button class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white">
            下载 <span class="hidden md:inline">蜃境客户端</span> macOS TestFlight版
          </button>
        </a>
        <p class="mt-2 text-sm text-gray-600">要求macOS High Sierra 10.13 及以上版本</p>
      </div>
    </div>
    <div class="mx-auto max-w-xl">
      <header class="my-6 pt-4">
        <h2 class="text-xl font-medium">在macOS系统上使用蜃境</h2>
      </header>
      <div class="Markdown mb-4">
        <ol>
          <li>安装并运行macOS应用</li>
          <li>接受任何安装VPN配置的提示</li>
          <li>
            从菜单栏上的蜃境图标菜单中点击<strong>登录</strong>并在弹出的浏览器中进行认证
          </li>
          <li>使用你或你团队使用的第三方服务进行认证，会自动完成注册登录</li>
        </ol>
      </div>
      <footer>
        <div class="Markdown">
          <p>
            你应该已经登录并连接上了！设置更多的设备接入你的网络，或者使用
            <a href="/admin" target="_blank" rel="noreferer noopener">控制台</a>
            来管理已有设备
          </p>
        </div>
      </footer>
    </div>
  </div>
</template>
