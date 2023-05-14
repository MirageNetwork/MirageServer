<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch, watchEffect } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate } from "vue-router";
import UserMenu from "./UserMenu.vue";
import ChangeRole from "./umenu/ChangeRole.vue";
import RemoveUser from "./umenu/RemoveUser.vue";
import Toast from "./Toast.vue";
import axios from "axios";

//与框架交互部分

//界面控制部分
const activeBtn = ref(null);
const btnLeft = ref(0);
const btnTop = ref(0);
function refreshUserMenuPos() {
  if (activeBtn.value != null) {
    btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14;
    btnTop.value = activeBtn.value?.getBoundingClientRect().top;
  }
}
function openUserMenu(u, event) {
  activeBtn.value = event.target;
  while (activeBtn.value?.tagName != "DIV" && activeBtn.value?.tagName != "div") {
    activeBtn.value = activeBtn.value.parentNode;
  }
  selectUser.value = u;
  btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14;
  btnTop.value = activeBtn.value?.getBoundingClientRect().top;
  userMenuShow.value = true;
}
function closeUserMenu() {
  activeBtn.value = null;
  userMenuShow.value = false;
}

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

const currentUserId = ref(0);
const ownerId = ref(-1);

const selectUser = ref({});
function mouseOnUser(u) {
  selectUser.value = u;
  userBtnShow.value = true;
}
function mouseLeaveUser() {
  userBtnShow.value = false;
}

const userMenuShow = ref(false);
const userBtnShow = ref(false);

const changeRoleShow = ref(false);
function showChangeRole() {
  userBtnShow.value = false;
  closeUserMenu();
  changeRoleShow.value = true;
}

const targetUserMList = ref([]);
const removeUserShow = ref(false);
function showRemoveUser() {
  userBtnShow.value = false;
  closeUserMenu();
  getMachines().then(function (mlist) {
    targetUserMList.value = [];
    for (var i in mlist) {
      if (mlist[i]["user"] == selectUser.value["loginName"]) {
        targetUserMList.value = targetUserMList.value.concat(mlist[i]);
      }
    }
    removeUserShow.value = true;
  });
}

const wantedRoles = ref({});
function setWantedRole(newWantedRole) {
  wantedRoles.value[selectUser.value["id"]] = newWantedRole;
}

function doChangeRole() {
  axios
    .post("/admin/api/users", {
      userID: selectUser.value["id"],
      action: "set_" + wantedRoles.value[selectUser.value["id"]],
    })
    .then(function (response) {
      if (response.data["status"] != "success") {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      } else {
        let newRoleChn = "普通成员";
        switch (wantedRoles.value[selectUser.value["id"]]) {
          case "admin":
            newRoleChn = "管理员";
            break;
          case "owner":
            newRoleChn = "所有者";
            break;
        }
        changeRoleShow.value = false;
        toastMsg.value =
          "已修改 " + selectUser.value["loginName"] + " 角色为 " + newRoleChn;
        toastShow.value = true;
        if (newRoleChn != "所有者") {
          getUsers().then().catch();
        }
      }
    })
    .catch(function (error) {
      toastMsg.value = error;
      toastShow.value = true;
    });
}

function doRemoveUser() {
  axios
    .post("/admin/api/users", {
      userID: selectUser.value["id"],
      action: "delete_user",
    })
    .then(function (response) {
      if (response.data["status"] != "success") {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      } else {
        removeUserShow.value = false;
        toastMsg.value = "已删除 " + selectUser.value["loginName"];
        toastShow.value = true;
        getUsers().then().catch();
      }
    })
    .catch(function (error) {
      toastMsg.value = error;
      toastShow.value = true;
    });
}

//数据填充控制部分
const UserList = ref({});
const usersNum = computed(() => {
  return UserList.value.length;
});
let getUserIntID;
function getUsers() {
  return new Promise((resolve, reject) => {
    axios
      .get("/admin/api/users")
      .then(function (response) {
        if (response.data["status"] != "success") {
          toastMsg.value = "获用户信息出错：" + response.data["status"].substring(6);
          toastShow.value = true;
          reject();
        }

        // 处理成功情况
        currentUserId.value = response.data["data"]["currentUserID"];
        ownerId.value = response.data["data"]["ownerID"];
        UserList.value = response.data["data"]["users"];
        for (let i in UserList.value) {
          wantedRoles.value[UserList.value[i]["id"]] = wantedRoles.value[
            UserList.value[i]["id"]
          ]
            ? wantedRoles.value[UserList.value[i]["id"]]
            : UserList.value[i]["role"];
        }
        resolve();
      })
      .catch(function (error) {
        // 处理错误情况
        toastMsg.value = "获取用户信息出错：" + error;
        toastShow.value = true;
        reject();
      });
  });
}
onMounted(() => {
  refreshUserMenuPos();
  window.addEventListener("resize", refreshUserMenuPos);
  window.addEventListener("scroll", refreshUserMenuPos);

  getUsers().then().catch();
  getUserIntID = setInterval(() => {
    getUsers().then().catch();
  }, 15000);
});

onUnmounted(() => {
  window.removeEventListener("resize", refreshUserMenuPos);
  window.removeEventListener("scroll", refreshUserMenuPos);
});
onBeforeRouteLeave(() => {
  clearInterval(getUserIntID);
});

function getMachines() {
  return new Promise((resolve, reject) => {
    axios
      .get("/admin/api/machines")
      .then(function (response) {
        if (
          response.data["needreauth"] != undefined ||
          response.data["needreauth"] == true
        ) {
          toastMsg.value =
            response.data["needreauthreason"] + "，登录状态失效，请重新登录";
          toastShow.value = true;
          reject();
        }
        // 处理成功情况
        if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
          resolve(response.data["mlist"]);
        } else if (response.data["errormsg"] != undefined) {
          toastMsg.value = "获取设备信息出错：" + response.data["errormsg"];
          toastShow.value = true;
          reject();
        }
      })
      .catch(function (error) {
        // 处理错误情况
        toastMsg.value = "获取设备信息出错：" + error;
        toastShow.value = true;
        reject();
      });
  });
}
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="mb-24">
      <header class="mb-8">
        <div class="flex justify-between items-center">
          <div class="flex items-center">
            <h1 class="text-3xl font-semibold tracking-tight leading-tight mb-2">用户</h1>
          </div>
        </div>
        <p class="text-gray-600">管理你网络中的用户和他们的权限</p>
      </header>

      <div
        class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-full px-2 py-1 leading-none text-sm mb-8"
      >
        {{ usersNum }} 个用户
      </div>
      <table class="table w-full">
        <thead>
          <tr>
            <th class="flex-auto table-cell items-center">用户</th>
            <th class="table-cell items-center md:w-1/4 lg:w-1/5">角色</th>
            <th class="hidden lg:table-cell items-center lg:w-1/5">加入日期</th>
            <th class="hidden lg:table-cell items-center lg:w-1/5">最近连线</th>
            <th class="table-cell justify-end ml-auto md:ml-0 relative items-center w-8">
              <span class="sr-only">用户操作菜单</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(u, id) in UserList">
            <tr
              :id="id"
              :v-if="u != nil"
              @mouseenter="mouseOnUser(u)"
              @mouseleave="mouseLeaveUser()"
              class="w-full px-0.5 hover"
            >
              <td class="flex-auto flex items-center">
                <div
                  class="relative shrink-0 rounded-full overflow-hidden transition-all duration-300 w-8 h-8 md:w-12 md:h-12 md:text-xl mr-3"
                >
                  <div
                    class="flex items-center justify-center text-center capitalize text-white font-medium pointer-events-none transition-all duration-300 w-8 h-8 md:w-12 md:h-12 md:text-xl"
                    style="background-color: rgb(161, 56, 33)"
                  >
                    {{ u.displayName[0] }}
                  </div>
                </div>
                <router-link class="relative" :to="'/machines?q=' + u.loginName">
                  <div class="items-center text-gray-900">
                    <p class="font-semibold hover:text-blue-500">
                      <a class="stretched-link">{{ u.displayName }} </a>
                    </p>
                    <span v-if="u.status == 'suspend'">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-red-50 bg-red-50 text-red-600 rounded-sm px-1 text-xs mr-1"
                      >
                        已冻结
                      </div>
                    </span>
                  </div>
                  <div class="flex items-center text-gray-600 text-sm">
                    <span>{{ u.loginName }} </span>
                  </div>
                </router-link>
              </td>
              <td class="table-cell items-center md:w-1/4 lg:w-1/5">
                <div class="flex relative min-w-0">
                  <div class="truncate">
                    <span>{{ u.role == "owner" ? "所有者" : "普通成员" }} </span>
                  </div>
                </div>
              </td>
              <td class="hidden lg:table-cell items-center lg:w-1/5">
                <time :datetime="u.created" :title="u.created">{{
                  new Date(u.created)
                    .toLocaleDateString()
                    .replace("/", "年")
                    .replace("/", "月") + "日"
                }}</time>
              </td>
              <td class="hidden lg:table-cell items-center lg:w-1/5">
                <span>
                  <div class="inline-flex items-center cursor-default">
                    <span
                      class="inline-block w-2 h-2 rounded-full mr-2"
                      :class="{
                        'bg-green-500': u.currentlyConnected,
                        'bg-gray-300': !u.currentlyConnected,
                      }"
                    ></span>
                    <span
                      v-if="u.currentlyConnected"
                      class="text-sm text-gray-600 tooltip tooltip-top"
                      :data-tip="
                        '最近在线于' +
                        new Date(u.lastSeen)
                          .toLocaleString()
                          .replace('/', '年')
                          .replace('/', '月')
                          .replace(' ', '日 ')
                      "
                      >已连接</span
                    >
                    <span
                      v-else
                      class="text-sm text-gray-600 tooltip tooltip-top"
                      :data-tip="
                        '最近在线于' +
                        new Date(u.lastSeen)
                          .toLocaleString()
                          .replace('/', '年')
                          .replace('/', '月')
                          .replace(' ', '日 ')
                      "
                      >{{
                        new Date(u.lastSeen)
                          .toLocaleString()
                          .replace("/", "年")
                          .replace("/", "月")
                          .replace(" ", "日 ")
                      }}
                    </span>
                  </div>
                </span>
              </td>
              <td
                class="table-cell justify-end ml-auto md:ml-0 relative items-center w-8"
              >
                <div
                  v-if="(!userBtnShow && !userMenuShow) || selectUser.id != u.id"
                  @click="openUserMenu(u, $event)"
                  class="flex-none w-12 -mt-0.5 relative"
                >
                  <button
                    class="py-0.5 px-2 shadow-none rounded-md border border-gray-300/0 hover:border-gray-300/100 hover:bg-gray-100 hover:shadow-md hover:cursor-pointer active:border-gray-300/100 active:shadow focus:outline-none focus:ring transition-shadow duration-100 ease-in-out z-20"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="24"
                      height="24"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      class="text-gray-500"
                    >
                      <circle cx="12" cy="12" r="1"></circle>
                      <circle cx="19" cy="12" r="1"></circle>
                      <circle cx="5" cy="12" r="1"></circle>
                    </svg>
                  </button>
                </div>
                <!---->
                <div
                  v-if="(userBtnShow || userMenuShow) && selectUser.id == u.id"
                  @click="openUserMenu(u, $event)"
                  class="flex-none w-12 border button-outline bg-white shadow-md cursor-pointer focus:outline-none focus:ring -mt-0.5 relative py-0.5 px-2 rounded-md border-gray-300/100 hover:border-gray-300/100 hover:bg-gray-100 hover:shadow-md hover:cursor-pointer active:border-gray-300/100 transition-shadow duration-100 ease-in-out z-20"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="text-gray-500"
                  >
                    <circle cx="12" cy="12" r="1"></circle>
                    <circle cx="19" cy="12" r="1"></circle>
                    <circle cx="5" cy="12" r="1"></circle>
                  </svg>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </section>
  </main>

  <!-- 提示框显示 -->
  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>

  <!--设备配置菜单显示-->
  <Teleport to="body">
    <UserMenu
      v-if="userMenuShow"
      :toleft="btnLeft"
      :totop="btnTop"
      :login-name="selectUser.loginName"
      :cant-edit="selectUser.isOwner || selectUser.id == currentUserId"
      @close="closeUserMenu"
      @showdialog-changerole="showChangeRole"
      @showdialog-removeuser="showRemoveUser"
    ></UserMenu>
  </Teleport>

  <!-- 菜单弹出提示框显示 -->
  <Teleport to="body">
    <!-- 修改角色提示框显示 -->
    <ChangeRole
      v-if="changeRoleShow"
      :select-user="selectUser"
      :wanted-role="wantedRoles[selectUser.id]"
      :can-assign-owner="currentUserId == ownerId"
      @close="changeRoleShow = false"
      @set-wantedrole="setWantedRole"
      @change-role="doChangeRole"
    ></ChangeRole>
    <!-- 移除用户提示框显示 -->
    <RemoveUser
      v-if="removeUserShow"
      :select-user="selectUser"
      :user-machine-list="targetUserMList"
      @close="removeUserShow = false"
      @confirm-remove="doRemoveUser"
    >
    </RemoveUser>
  </Teleport>
</template>

<style scoped>
.table tr.hover:hover th,
.table tr.hover:hover td,
.table tr.hover:nth-child(even):hover th,
.table tr.hover:nth-child(even):hover td {
  background-color: #faf9f8;
}

.table :where(thead, tfoot) :where(th, td) {
  background-color: #ffffff;
  color: #71706f;
  border-bottom-width: 1px;
}

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
