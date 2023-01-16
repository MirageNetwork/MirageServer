<script setup>
import { ref, onMounted } from "vue";

//界面控制部分
const userMenuOpen = ref(false)

function switchUserMenu() {
  userMenuOpen.value = !userMenuOpen.value
  if (userMenuOpen.value) {
    document.body.style.pointerEvents = "none"
  } else {
    document.body.style.removeProperty("pointer-events")
  }
}

//数据填充控制部分
const UserAccount = ref("");
const Basedomain = ref("");
const UserName = ref("");
const UserNameHead = ref("");
onMounted(() => {
  axios
    .get("/admin/api/self")
    .then(function (response) {
      // 处理成功情况

      if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
        UserAccount.value = response.data["useraccount"];
        Basedomain.value = response.data["basedomain"];
        UserName.value = response.data["username"];
        UserNameHead.value = response.data["usernamehead"];
      }
    })
    .catch(function (error) {
      // 处理错误情况
      alert(error);
    })
    .then(function () {
      // 总是会执行
    });
});
</script>

<template>
  <div class="bg-gray-100 border-b border-gray-200 pt-4 mb-6">
    <div class="container mx-auto mb-4 md:mb-6">
      <header class="flex justify-between items-center px-2 md:px-0">
        <a href="/admin" class="flex items-center" style="max-width: 80%">
          <img width="18" height="18" src="/img/mlogo.png" />
          <div role="banner" class="text-lg font-semibold ml-3 truncate">
            {{ UserAccount }}.{{ Basedomain }}
          </div>
          <span class="badge badge-secondary">仅供测试</span>
        </a>

        <nav class="flex items-center">
          <a class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
            href="https://github.com/gps949/tailscale/releases" target="_blank" rel="noopener noreferrer">下载客户端</a>
          <div @blur="switchUserMenu" class="dropdown dropdown-open dropdown-end" tabindex="-1">
            <div @click="switchUserMenu" cursor="pointer" class="avatar placeholder">
              <div class="bg-neutral-focus text-neutral-content rounded-full w-8">
                <span class="text-xs"> {{ UserNameHead }} </span>
              </div>
            </div>
            <div v-if="userMenuOpen" class="menu dropdown-content py-1 shadow bg-base-100 rounded-box w-52 mt-4">
              <div class="dropdown bg-white rounded-md py-1 z-50"
                style="outline: none; --radix-dropdown-menu-content-transform-origin: var(--radix-popper-transform-origin); pointer-events: auto;">
                <div class="block px-4 py-2">
                  <strong> {{ UserName }} </strong> <br />
                  {{ UserAccount }}
                </div>
                <div class="my-1 border-b border-gray-200"></div>
                <div onclick="window.location.href='/admin/logout'"
                  class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                  登出
                </div>
              </div>
            </div>
          </div>
        </nav>
      </header>
    </div>
    <div class="relative overflow-hidden" style="top: 1px">
      <nav id="nav"
        class="navigation flex items-center overflow-auto left-1 relative md:container md:mx-auto md:px-0 md:-left-3">
        <router-link class="whitespace-nowrap py-2 group relative" permission="devices" to="/admin/machines">
          <div
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5 after:bg-blue-600 text-blue-600 after:visible">
            <svg xmlns="http://www.w3.org/2000/svg" width="1.125em" height="1.125em" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
              class="mr-2 hidden sm:inline-block">
              <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
              <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
              <line x1="6" y1="6" x2="6.01" y2="6"></line>
              <line x1="6" y1="18" x2="6.01" y2="18"></line>
            </svg>
            <div data-content="Machines" class="navigation-link navigation-linkActive">
              设备
            </div>
          </div>
        </router-link>
      </nav>
    </div>
  </div>
  <router-view></router-view>
</template>
