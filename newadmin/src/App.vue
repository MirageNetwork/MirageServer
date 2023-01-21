<script setup>
import { ref, onMounted, watch, watchEffect, computed } from "vue";
import { useRouter } from "vue-router";
import UserMenu from "./components/UserMenu.vue";

const router = useRouter();
//界面控制部分
const userAvatar = ref(null)
const avatarLeft = ref(0)
const avatarTop = ref(0)
function watchWindowChange() {
  avatarLeft.value = userAvatar.value?.getBoundingClientRect().left
  avatarTop.value = userAvatar.value?.getBoundingClientRect().top
  window.onresize = () => {
    avatarLeft.value = userAvatar.value?.getBoundingClientRect().left
    avatarTop.value = userAvatar.value?.getBoundingClientRect().top
  }
  window.onscroll = () => {
    avatarLeft.value = userAvatar.value?.getBoundingClientRect().left
    avatarTop.value = userAvatar.value?.getBoundingClientRect().top
  }
}

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
  watchWindowChange()

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
  <div class="bg-base-200 border-b border-base-300 pt-4 mb-6">
    <div class="container mx-auto mb-4 md:mb-6">
      <header class="flex justify-between items-center px-2 md:px-0">
        <a href="/" class="flex items-center" style="max-width: 80%">
          <img width="18" height="18" src="/img/mlogo.png" />
          <div role="banner" class="text-lg font-semibold ml-3 truncate">
            {{ UserAccount }}.{{ Basedomain }}
          </div>
          <span class="badge badge-secondary">仅供测试</span>
        </a>

        <nav class="flex items-center">
          <a class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
            href="https://github.com/gps949/tailscale/releases" target="_blank" rel="noopener noreferrer">下载客户端</a>


          <div ref="userAvatar" @click="switchUserMenu" class="heart-wrapper ml-2"><button
              class="relative rounded-full">
              <div class="relative shrink-0 rounded-full overflow-hidden w-8 h-8">
                <div
                  class="flex items-center justify-center text-center capitalize text-white font-medium pointer-events-none w-8 h-8"
                  style="background-color: rgb(161, 56, 33);">{{ UserNameHead }} </div>
              </div>
            </button></div>


          <!--
          <div  cursor="pointer" class="avatar placeholder">
            <div class="bg-neutral-focus text-neutral-content rounded-full w-8">
              <span class="text-xs"> {{ UserNameHead }} </span>
            </div>
          </div>
-->
          <!--用户菜单部分-->
          <Teleport to="body">
            <UserMenu v-if="userMenuOpen" :toleft="avatarLeft" :totop="avatarTop" :user-account="UserAccount"
              :user-name="UserName" @close="switchUserMenu"></UserMenu>
          </Teleport>
        </nav>
      </header>
    </div>
    <div class="relative overflow-hidden" style="top: 1px">
      <nav id="nav"
        class="navigation flex items-center overflow-auto left-1 relative md:container md:mx-auto md:px-0 md:-left-3">
        <router-link class="whitespace-nowrap py-2 group relative" to="/machines">
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
            <div class="navigation-link navigation-linkActive">
              设备
            </div>
          </div>
        </router-link>
      </nav>
    </div>
  </div>
  <router-view></router-view>
</template>
