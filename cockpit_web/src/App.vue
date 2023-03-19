<script setup>
import { ref, onMounted, computed, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import UserMenu from "./components/UserMenu.vue";

const router = useRouter();
const route = useRoute();

//界面控制部分
const userAvatar = ref(null);
const avatarLeft = ref(0);
const avatarTop = ref(0);

const needRegister = ref(false);
const needReauth = ref(false);
const netErrMsg = ref("");

function refreshUserMenuPos() {
  avatarLeft.value = userAvatar.value?.getBoundingClientRect().left;
  avatarTop.value = userAvatar.value?.getBoundingClientRect().top;
}

const userMenuOpen = ref(false);
function switchUserMenu() {
  refreshUserMenuPos();
  userMenuOpen.value = !userMenuOpen.value;
  if (userMenuOpen.value) {
    document.body.style.pointerEvents = "none";
  } else {
    document.body.style.removeProperty("pointer-events");
  }
}

const currentRoute = computed(() => {
  let curPath = route.path;
  if (curPath == "/") return "setting";
  if (curPath.substring(0, 9) == "/regAdmin") return "regAdmin";
  if (curPath.substring(0, 6) == "/login") return "login";
  if (curPath.substring(0, 6) == "/users") return "users";
  if (curPath.substring(0, 8) == "/orgs") return "orgs";
  if (curPath.substring(0, 8) == "/setting") return "setting";
});

const serviceSwitch = ref(null);
const serviceState = ref("stopped");
const serviceStateStr = {
  running: "控制器正运行",
  stopped: "控制器已停止",
  starting: "控制器启动中",
  stopping: "控制器停止中",
};

watch(
  () => serviceState.value,
  (newVal) => {
    switch (newVal) {
      case "running":
        serviceSwitch.value.indeterminate = false;
        serviceSwitch.value.checked = true;
        break;
      case "stopped":
        serviceSwitch.value.indeterminate = false;
        serviceSwitch.value.checked = false;
        break;
      default:
        serviceSwitch.value.indeterminate = true;
        break;
    }
  }
);
function doServiceSwitch() {
  if (serviceSwitch.value.checked == false && serviceState.value == "running") {
    serviceState.value = "stopping";
    axios
      .post("/cockpit/api/service/stop")
      .then((res) => {
        serviceState.value = res.data["data"] ? "running" : "stopped";
      })
      .catch((err) => {
        console.log(err);
      });
  } else if (serviceSwitch.value.checked == true && serviceState.value == "stopped") {
    serviceState.value = "starting";
    axios
      .post("/cockpit/api/service/start")
      .then((res) => {
        serviceState.value = res.data["data"] ? "running" : "stopped";
      })
      .catch((err) => {
        console.log(err);
      });
  }
}

function getServiceState() {
  axios
    .get("/cockpit/api/service/state")
    .then((res) => {
      serviceState.value = res.data["data"] ? "running" : "stopped";
    })
    .catch((err) => {
      console.log(err);
    });
}

onMounted(() => {
  window.addEventListener("resize", refreshUserMenuPos);
  window.addEventListener("scroll", refreshUserMenuPos);
  axios.interceptors.response.use(
    (response) => {
      if (response.status == 200) {
        if (response.data["status"] == "error-noadmin") {
          needRegister.value = true;
          netErrMsg.value = "未绑定超级管理员";
          return response;
        }
        if (response.data["status"] == "error-unauthorized") {
          needRegister.value = false;
          needReauth.value = true;
          netErrMsg.value = "未登录或登录状态超时失效";
          return response;
        }
        needRegister.value = false;
        netErrMsg.value = "";
        needReauth.value = false;
      } else {
        needRegister.value = false;
        netErrMsg.value = "登录状态超时失效";
        needReauth.value = true;
      }
      return response;
    },
    (error) => {
      needRegister.value = false;
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
  getServiceState().then().catch();
  getServiceStateIntID = setInterval(() => {
    getServiceState().then().catch();
  }, 5000);
});
function logoutDone() {
  netErrMsg.value = "未登录或登录状态超时失效";
  needReauth.value = true;
  router.push("/login");
}
function doLogout() {
  axios
    .get("/cockpit/api/logout")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        logoutDone();
        return;
      }
      console.log("Get Logout Response ", response.data["status"]);
    })
    .catch(function (error) {
      // 处理错误情况
      console.log(error);
    });
}
</script>

<template>
  <div
    v-if="needRegister && currentRoute != 'regAdmin'"
    class="bg-amber-700 text-white font-medium py-2 px-4 text-center"
  >
    尚无超级管理员，请
    <router-link class="text-amber-100" to="/regAdmin">绑定超级管理员</router-link>
  </div>
  <div
    v-if="needReauth && currentRoute != 'login'"
    class="bg-amber-700 text-white font-medium py-2 px-4 text-center"
  >
    {{ netErrMsg }}，请
    <router-link class="text-amber-100" to="/login">前往登录</router-link>
  </div>
  <div class="bg-base-200 border-b border-base-300 pt-4 mb-6">
    <div class="container mx-auto mb-4 md:mb-6">
      <header class="flex justify-between items-center px-2 md:px-0">
        <div class="flex items-center">
          <a href="/cockpit" class="flex items-center" style="max-width: 80%">
            <img width="18" height="18" src="/img/logo.svg" />
            <div role="banner" class="text-lg font-semibold ml-3 truncate min-w-fit">
              蜃境系统管理
            </div>
          </a>
        </div>

        <nav v-if="!needReauth && !needRegister" class="flex items-center">
          <span
            v-if="!needReauth && !needRegister"
            class="pl-0 border-l-0 text-stone-200 badge badge-lg"
            :class="{
              'badge-success': serviceState == 'running',
              'badge-error': serviceState == 'stopped',
              'badge-warning': serviceState == 'starting' || serviceState == 'stopping',
            }"
          >
            <input
              @change="doServiceSwitch"
              ref="serviceSwitch"
              type="checkbox"
              class="toggle toggle-md mr-2"
              :class="{
                'toggle-success': serviceState == 'running',
                'toggle-error': serviceState == 'stopped',
                'toggle-warning':
                  serviceState == 'starting' || serviceState == 'stopping',
              }"
            />
            {{ serviceStateStr[serviceState] }}
          </span>
          <button
            ref="userAvatar"
            @click="doLogout"
            class="relative rounded-full overflow-hidden ml-7 w-7 h-7"
          >
            <div class="flex items-center justify-center pointer-events-none">
              <svg
                t="1679192483610"
                class="icon"
                viewBox="0 0 1024 1024"
                version="1.1"
                xmlns="http://www.w3.org/2000/svg"
                p-id="5172"
                width="28"
                height="28"
              >
                <path
                  d="M469.333333 128a383.018667 383.018667 0 0 1 285.44 127.146667l-95.146666 85.632a256 256 0 1 0 0 342.485333l95.146666 85.632A384 384 0 1 1 469.333333 128z m256 213.333333l234.666667 170.666667-234.666667 170.666667v-106.666667h-256v-128h256V341.333333z"
                  fill="#252B2F"
                  p-id="5173"
                ></path>
              </svg>
            </div>
          </button>
          <!--用户菜单部分-->
          <Teleport to="body">
            <UserMenu
              v-if="userMenuOpen"
              :toleft="avatarLeft"
              :totop="avatarTop"
              user-account="SuperAdmin"
              user-name="蜃境超级管理员"
              @close="switchUserMenu"
              @logout="logoutDone"
            ></UserMenu>
          </Teleport>
        </nav>
      </header>
    </div>
    <div class="relative overflow-hidden">
      <nav
        v-if="!needReauth && !needRegister"
        id="nav"
        class="navigation flex items-center overflow-auto left-1 relative md:container md:mx-auto md:px-0 md:-left-3"
      >
        <router-link class="whitespace-nowrap py-2 group relative" to="/orgs">
          <div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'orgs',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'orgs',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5 after:bg-blue-600"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'orgs' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
              <polyline points="9 22 9 12 15 12 15 22"></polyline>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'orgs' }">租户</div>
          </div>
        </router-link>

        <router-link class="whitespace-nowrap py-2 group relative" to="/users">
          <div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'users',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'users',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5 after:bg-blue-600"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentRoute == 'users' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="mr-2 inline-block"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <div :class="{ 'font-medium': currentRoute == 'users' }">用户</div>
          </div>
        </router-link>

        <router-link class="whitespace-nowrap py-2 group relative" to="/setting">
          <div
            :class="{
              'text-blue-600 after:visible': currentRoute == 'setting',
              'text-gray-600 group-hover:text-gray-800 after:invisible':
                currentRoute != 'setting',
            }"
            class="px-3 py-2 flex items-center rounded-md group-hover:bg-gray-200 after:absolute after:bottom-0 after:right-3 after:left-3 after:h-0.5 after:bg-blue-600"
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
  <router-view
    v-if="
      ((currentRoute == 'login' || !needReauth) && !needRegister) ||
      currentRoute == 'regAdmin'
    "
  ></router-view>
  <main
    v-if="
      (needReauth && currentRoute != 'login') ||
      (needRegister && currentRoute != 'regAdmin')
    "
    class="container mx-auto pb-20 md:pb-24"
  >
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
          <div><strong>错误：</strong> {{ netErrMsg }}</div>
        </div>
      </div>
    </section>
  </main>
</template>

<style>
.container {
  width: 94%;
}
/*
.toggle {
  border: 0;
  --tglbg: #d6d3d1;
  background-color: white;
}

.toggle:checked {
  border: 0;
  --tglbg: #1e40af;
  background-color: white;
}

.toggle:disabled {
  --togglehandleborder: 0 0 0 3px white inset,
    var(--handleoffsetcalculator) 0 0 3px white inset;
}
*/
</style>
