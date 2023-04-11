<script setup>
import { ref, onMounted, watch, watchEffect, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import SelfMenu from "./components/SelfMenu.vue";

const router = useRouter();
const route = useRoute();

//界面控制部分
const selfAvatar = ref(null);
const avatarLeft = ref(0);
const avatarTop = ref(0);

const needReauth = ref(false);
const netErrMsg = ref("");
const curURL = computed(() => {
  return route.path;
});

function refreshSelfMenuPos() {
  avatarLeft.value = selfAvatar.value?.getBoundingClientRect().left;
  avatarTop.value = selfAvatar.value?.getBoundingClientRect().top;
}

const selfMenuOpen = ref(false);
function switchSelfMenu() {
  selfMenuOpen.value = !selfMenuOpen.value;
  if (selfMenuOpen.value) {
    document.body.style.pointerEvents = "none";
  } else {
    document.body.style.removeProperty("pointer-events");
  }
}

const currentRoute = computed(() => {
  let curPath = route.path;
  if (curPath == "/") return "machines";
  if (curPath.substring(0, 6) == "/users") return "users";
  if (curPath.substring(0, 4) == "/dns") return "dns";
  if (curPath.substring(0, 8) == "/machine") return "machines";
  if (curPath.substring(0, 8) == "/setting") return "setting";
  if (curPath.substring(0, 5) == "/acls") return "acls";
});

//数据填充控制部分
const SelfAccount = ref("");
const Basedomain = ref("");
const SelfName = ref("");
const SelfNameHead = ref("");
const OrgName = ref("");

let getSelfIntID;
function getSelf() {
  return new Promise((resolve, reject) => {
    axios
      .get("/admin/api/self")
      .then(function (response) {
        // 处理成功情况
        if (response.data["status"] == "success") {
          SelfAccount.value = response.data["data"]["useraccount"];
          Basedomain.value = response.data["data"]["basedomain"];
          SelfName.value = response.data["data"]["username"];
          SelfNameHead.value = response.data["data"]["usernamehead"];
          OrgName.value = response.data["data"]["orgname"];
          resolve("success");
        } else {
          reject("err");
        }
        reject("err");
      })
      .catch(function (error) {
        // 处理错误情况
        reject("error");
      });
  });
}

onMounted(() => {
  refreshSelfMenuPos();
  window.addEventListener("resize", refreshSelfMenuPos);
  window.addEventListener("scroll", refreshSelfMenuPos);

  axios.interceptors.response.use(
    (response) => {
      if (response.status == 200) {
        netErrMsg.value = "";
        needReauth.value = false;
      } else {
        netErrMsg.value = "登录状态超时失效";
        needReauth.value = true;
      }
      return response;
    },
    (error) => {
      if (error && error.response) {
        switch (error.response.status) {
          case 400:
            error.message = "请求错误(400)";
            break;
          case 401:
            error.message = "未授权，请重新登录(401)";
            break;
          case 403:
            error.message = "拒绝访问(403)";
            break;
          case 404:
            error.message = "请求出错(404)";
            break;
          case 408:
            error.message = "请求超时(408)";
            break;
          case 500:
            error.message = "服务器错误(500)";
            break;
          case 501:
            error.message = "服务未实现(501)";
            break;
          case 502:
            error.message = "网络错误(502)";
            break;
          case 503:
            error.message = "服务不可用(503)";
            break;
          case 504:
            error.message = "网络超时(504)";
            break;
          case 505:
            error.message = "HTTP版本不受支持(505)";
            break;
          default:
            error.message = "连接出错" + error.response.status;
        }
      } else {
        error.message = "连接服务器失败!";
      }
      netErrMsg.value = error.message;
      needReauth.value = true;
      return Promise.reject(error);
    }
  );

  getSelf().then().catch();
  getSelfIntID = setInterval(() => {
    getSelf().then().catch();
  }, 15000);
});
</script>

<template>
  <div
    v-if="needReauth"
    class="bg-amber-700 text-white font-medium py-2 px-4 text-center"
  >
    连接服务器出现{{ netErrMsg }}，请尝试
    <a class="text-amber-100" :href="'/login?next_url=/admin#' + curURL">重新登录</a>
  </div>
  <div class="bg-base-200 border-b border-base-300 pt-4 mb-6">
    <div class="container mx-auto mb-4 md:mb-6">
      <header class="flex justify-between items-center px-2 md:px-0">
        <a href="/" class="flex items-center" style="max-width: 80%">
          <img width="18" height="18" src="/img/logo.svg" />
          <div role="banner" class="text-lg font-semibold ml-3 truncate">
            {{ OrgName }}
          </div>
        </a>

        <nav class="flex items-center">
          <a
            class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
            href="https://github.com/gps949/tailscale/releases"
            target="_blank"
            rel="noopener noreferrer"
            >下载客户端</a
          >

          <div ref="selfAvatar" @click="switchSelfMenu" class="heart-wrapper ml-2">
            <button class="relative rounded-full">
              <div class="relative shrink-0 rounded-full overflow-hidden w-8 h-8">
                <div
                  class="flex items-center justify-center text-center capitalize text-white font-medium pointer-events-none w-8 h-8"
                  style="background-color: rgb(161, 56, 33)"
                >
                  {{ SelfNameHead }}
                </div>
              </div>
            </button>
          </div>

          <!--用户菜单部分-->
          <Teleport to="body">
            <SelfMenu
              v-if="selfMenuOpen"
              :toleft="avatarLeft"
              :totop="avatarTop"
              :self-account="SelfAccount"
              :self-name="SelfName"
              @close="switchSelfMenu"
            ></SelfMenu>
          </Teleport>
        </nav>
      </header>
    </div>
    <div class="relative overflow-hidden" style="top: 1px">
      <nav
        id="nav"
        class="navigation flex items-center overflow-auto left-1 relative md:container md:mx-auto md:px-0 md:-left-3"
      >
        <router-link class="whitespace-nowrap py-2 group relative" to="/machines">
          <div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'machines',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'machines',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1.125em"
              height="1.125em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'machines' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
              <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
              <line x1="6" y1="6" x2="6.01" y2="6"></line>
              <line x1="6" y1="18" x2="6.01" y2="18"></line>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'machines' }">设备</div>
          </div>
        </router-link>

        <router-link class="whitespace-nowrap py-2 group relative" to="/users"
          ><div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'users',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'users',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1.125em"
              height="1.125em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'users' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'users' }">用户</div>
          </div>
        </router-link>

        <router-link class="whitespace-nowrap py-2 group relative" to="/acls"
          ><div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'acls',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'acls',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1.125em"
              height="1.125em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'acls' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'acls' }">访问控制</div>
          </div>
        </router-link>

        <router-link class="whitespace-nowrap py-2 group relative" to="/dns">
          <div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'dns',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'dns',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1.125em"
              height="1.125em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'dns' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="2" y1="12" x2="22" y2="12"></line>
              <path
                d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
              ></path>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'dns' }">DNS</div>
          </div>
        </router-link>

        <router-link class="whitespace-nowrap py-2 group relative" to="/settings">
          <div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'setting',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'setting',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1.125em"
              height="1.125em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'setting' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <circle cx="12" cy="12" r="3"></circle>
              <path
                d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
              ></path>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'setting' }">设置</div>
          </div>
        </router-link>
      </nav>
    </div>
  </div>
  <router-view v-if="!needReauth"></router-view>
  <main v-if="needReauth" class="container mx-auto pb-20 md:pb-24">
    <section class="mb-24">
      <div class="w-full p-3 flex items-center justify-center text-sm">
        <div class="flex items-center justify-center">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="mr-3 text-red-400 h-5 w-5"
          >
            <path
              d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
            ></path>
            <line x1="12" y1="9" x2="12" y2="13"></line>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
          <div><strong>错误：</strong> 请求失败 {{ netErrMsg }}</div>
        </div>
      </div>
    </section>
  </main>
</template>

<style>
.container {
  width: 94%;
}
</style>
