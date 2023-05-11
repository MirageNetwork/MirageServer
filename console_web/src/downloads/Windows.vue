<script setup>
const props = defineProps({
  downloadDetails: Object,
});
</script>
<template>
  <div>
    <div class="pt-4 pb-10 border-b border-gray-200 max-w-xl mx-auto">
      <div v-if="downloadDetails.windows.primary == ''" class="text-center">
        <button
          disabled
          class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white disabled:bg-blue-600/60 disabled:text-white/60"
        >
          暂未提供Windows版本
        </button>
      </div>
      <div v-if="downloadDetails.windows.primary != ''" class="text-center">
        <a :href="downloadDetails.windows.primary">
          <button class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white">
            下载 <span class="hidden md:inline">蜃境客户端</span> Windows版
          </button>
        </a>
        <p class="mt-2 text-sm text-gray-600">要求 Windows 7 SP1 及以上版本</p>
      </div>
    </div>
    <div class="mx-auto max-w-xl">
      <header class="my-6 pt-4">
        <h2 class="text-xl font-medium">在Windows系统上使用蜃境</h2>
      </header>
      <div class="Markdown mb-4">
        <ol>
          <li>下载并运行Windows安装程序</li>
          <li>
            运行蜃境，右键点击你系统托盘中蜃境图标，弹出菜单中单击 <strong>登录</strong>
            并从弹出的浏览器窗口中完成认证。（过程中您可能需要根据提示设置控制器地址）
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
