<script setup>
import vueQr from "vue-qr/src/packages/vue-qr.vue";
const props = defineProps({
  downloadDetails: Object,
});
</script>
<template>
  <div>
    <div class="pt-4 pb-10 border-b border-gray-200 max-w-xl mx-auto">
      <div
        v-if="downloadDetails.ios.primary == '' && downloadDetails.ios.secondary == ''"
        class="text-center"
      >
        <button
          disabled
          class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white disabled:bg-blue-600/60 disabled:text-white/60"
        >
          暂未提供iOS版本
        </button>
      </div>
      <div
        v-if="downloadDetails.ios.primary != ''"
        class="md:flex items-center justify-center"
      >
        <a
          :href="downloadDetails.ios.primary"
          class="border overflow-hidden rounded-lg block max-w-[10rem] mx-auto md:mx-0 mb-8 md:mb-0"
          data-tooltip="使用你手机的相机扫描QR码下载蜃境客户端"
        >
          <vue-qr :text="downloadDetails.ios.primary" :size="400"></vue-qr>
        </a>
        <div class="max-w-[18rem] md:ml-6 text-center md:text-left mx-auto md:mx-0">
          <p class="mb-3 text-gray-600">扫描二维码或用下方链接下载蜃境iOS版</p>
          <a :href="downloadDetails.ios.primary"
            ><button class="btn btn-sm border-0 bg-blue-600 hover:bg-blue-700 text-white">
              下载 <span class="hidden md:inline">蜃境客户端</span> iOS版
            </button>
          </a>
          <p class="mt-2 text-sm text-gray-600">需要iOS 15及以上系统</p>
        </div>
      </div>
      <div
        v-if="downloadDetails.ios.primary == '' && downloadDetails.ios.secondary != ''"
        class="md:flex items-center justify-center"
      >
        <a
          :href="downloadDetails.ios.secondary"
          class="border overflow-hidden rounded-lg block max-w-[10rem] mx-auto md:mx-0 mb-8 md:mb-0"
          data-tooltip="使用你手机的相机扫描QR码下载蜃境客户端"
        >
          <vue-qr :text="downloadDetails.ios.secondary" :size="400"></vue-qr>
        </a>
        <div class="max-w-[18rem] md:ml-6 text-center md:text-left mx-auto md:mx-0">
          <p class="mb-3 text-gray-600">扫描二维码或用下方链接下载蜃境iOS TestFlight版</p>
          <a :href="downloadDetails.ios.secondary"
            ><button class="btn btn-sm border-0 bg-blue-600 hover:bg-blue-700 text-white">
              下载 <span class="hidden md:inline">蜃境客户端</span> iOS TestFlight版
            </button>
          </a>
          <p class="mt-2 text-sm text-gray-600">需要iOS 15及以上系统</p>
        </div>
      </div>
    </div>
    <div class="mx-auto max-w-xl">
      <header class="my-6 pt-4">
        <h2 class="text-xl font-medium">在iOS系统上使用蜃境</h2>
      </header>
      <div class="Markdown mb-4">
        <ol>
          <li>安装并打开蜃境iOS App</li>
          <li>根据屏幕提示，同意安装VPN配置，并使用右上角菜单设置控制器地址</li>
          <li>
            点击<strong>登录</strong>使用你或你团队使用的第三方服务进行认证，会自动完成注册登录
          </li>
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
