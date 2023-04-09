<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch, watchEffect } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate } from "vue-router";
import NaviNodeMenu from "./components/NaviNodeMenu.vue";

import Deploy from "./derp/Deploy.vue";
import RemoveNavi from "./derp/RemoveNavi.vue";
import NaviDetails from "./derp/NaviDetails.vue";
import Toast from "./components/Toast.vue";

//与框架交互部分

//界面控制部分
const activeBtn = ref(null);
const btnLeft = ref(0);
const btnTop = ref(0);
function refreshNaviNodeMenuPos() {
  if (activeBtn.value != null) {
    btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14;
    btnTop.value = activeBtn.value?.getBoundingClientRect().top;
  }
}
function openNaviNodeMenu(nr, nn, event) {
  activeBtn.value = event.target;
  while (activeBtn.value?.tagName != "DIV" && activeBtn.value?.tagName != "div") {
    activeBtn.value = activeBtn.value.parentNode;
  }
  selectNaviNode.value = nn;
  btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14;
  btnTop.value = activeBtn.value?.getBoundingClientRect().top;
  NaviNodeMenuShow.value = true;
}
function closeNaviNodeMenu() {
  activeBtn.value = null;
  NaviNodeMenuShow.value = false;
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

const selectNaviNode = ref({});
function mouseOnNaviNode(u) {
  selectNaviNode.value = u;
  NaviNodeBtnShow.value = true;
}
function mouseLeaveNaviNode() {
  NaviNodeBtnShow.value = false;
}

const NaviNodeMenuShow = ref(false);
const NaviNodeBtnShow = ref(false);

const removeNaviShow = ref(false);
function showRemoveNavi() {
  NaviNodeBtnShow.value = false;
  closeNaviNodeMenu();
  removeNaviShow.value = true;
}

const naviDetailsShow = ref(false);
function showNaviDetails() {
  NaviNodeBtnShow.value = false;
  closeNaviNodeMenu();
  naviDetailsShow.value = true;
}

const deployDERPShow = ref(false);
function showDeployDERP() {
  deployDERPShow.value = true;
}

function addNaviDone(newlist) {
  toastShow.value = true;
  toastMsg.value = "添加成功";
  NaviRegionList.value = newlist;
  deployDERPShow.value = false;
}

function doRemoveNavi() {
  axios
    .delete("/cockpit/api/derp/" + selectNaviNode.value["Name"], {})
    .then(function (response) {
      if (response.data["status"] == "success") {
        removeNaviShow.value = false;
        toastMsg.value = "已删除 " + selectNaviNode.value["HostName"];
        toastShow.value = true;
        NaviRegionList.value = response.data["data"];
      } else {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      toastMsg.value = error;
      toastShow.value = true;
    });
}

//数据填充控制部分
const NaviRegionList = ref([]);
const NaviRegionNum = computed(() => {
  if (NaviRegionList.value == null) {
    return 0;
  }
  return NaviRegionList.value.length;
});
let getNaviRegionsIntID;
function getNaviRegions() {
  return new Promise((resolve, reject) => {
    axios
      .get("/cockpit/api/derp/query")
      .then(function (response) {
        if (response.data["status"] != "success") {
          toastMsg.value = "获租户信息出错：" + response.data["status"].substring(6);
          toastShow.value = true;
          reject();
        }

        // 处理成功情况
        NaviRegionList.value = response.data["data"];
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
  refreshNaviNodeMenuPos();
  window.addEventListener("resize", refreshNaviNodeMenuPos);
  window.addEventListener("scroll", refreshNaviNodeMenuPos);

  getNaviRegions().then().catch();
  getNaviRegionsIntID = setInterval(() => {
    getNaviRegions().then().catch();
  }, 20000);
});

onUnmounted(() => {
  window.removeEventListener("resize", refreshNaviNodeMenuPos);
  window.removeEventListener("scroll", refreshNaviNodeMenuPos);
});

onBeforeRouteLeave(() => {
  clearInterval(getNaviRegionsIntID);
});

function secondsFormat(s) {
  var day = Math.floor(s / (24 * 3600)); // Math.floor()向下取整
  var hour = Math.floor((s - day * 24 * 3600) / 3600);
  var minute = Math.floor((s - day * 24 * 3600 - hour * 3600) / 60);
  var second = s - day * 24 * 3600 - hour * 3600 - minute * 60;
  return (
    (day > 0 ? day + "天" : "") +
    (hour > 0 ? hour + "小时" : "") +
    (minute > 0 ? minute + "分" : "") +
    second +
    "秒"
  );
}
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="mb-24">
      <header class="mb-4 flex items-center">
        <div class="flex justify-between items-center min-w-fit">
          <div class="flex items-center">
            <h1 class="text-3xl font-semibold tracking-tight leading-tight mb-2">司南</h1>
          </div>
        </div>
        <div
          class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-full px-2 py-1 leading-none text-sm ml-4 min-w-fit h-7"
        >
          {{ NaviRegionNum }} 个区域
        </div>
        <div class="flex w-full justify-end">
          <input
            type="button"
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
            value="添加新司南"
            @click="showDeployDERP"
          />
        </div>
      </header>

      <template v-for="nr in NaviRegionList">
        <div
          class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-full px-2 py-1 leading-none text-sm ml-4 min-w-fit h-7"
        >
          {{ nr.Region.RegionID }} 号区-{{ nr.Region.RegionCode }}-{{
            nr.Region.RegionName
          }}
          共 {{ nr.Nodes ? nr.Nodes.length : 0 }} 只司南
        </div>
        <table class="table w-full mb-3">
          <thead>
            <tr>
              <th
                class="md:w-1/4 flex-auto md:flex-initial md:shrink-0 w-0 text-ellipsis pt-2 pb-1"
              >
                司南
              </th>
              <th class="hidden md:table-cell md:w-1/4 pt-2 pb-1">IP</th>
              <th class="hidden md:table-cell w-1/4 lg:w-1/5 pt-2 pb-1">端口</th>
              <th class="hidden lg:table-cell md:flex-auto pt-2 pb-1">状态</th>
              <th
                class="table-cell justify-end ml-auto md:ml-0 relative w-1/6 lg:w-12 pt-2 pb-1"
              >
                <span class="sr-only">司南操作菜单</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <template v-for="nn in nr.Nodes">
              <tr
                :v-if="nn != nil"
                @mouseenter="mouseOnNaviNode(nn)"
                @mouseleave="mouseLeaveNaviNode()"
                class="w-full px-0.5 hover"
              >
                <td
                  class="md:w-1/4 flex-auto md:flex-initial md:shrink-0 w-0 text-ellipsis"
                >
                  <div class="relative">
                    <div class="items-center text-gray-900">
                      <p class="font-semibold hover:text-blue-500">
                        <span
                          :class="{
                            'bg-green-500': nn.Statics.latency != -1,
                            'bg-gray-300': nn.Statics.latency == -1,
                          }"
                          class="inline-block w-2 h-2 rounded-full relative -top-px lg:hidden mr-2"
                        ></span>
                        <a class="stretched-link">{{ nn.HostName }} </a>
                        <span v-if="nn.Arch == 'external'" class="ml-1">
                          <div
                            class="inline-flex items-center align-middle justify-center font-medium border border-red-50 bg-red-50 text-red-600 rounded-sm px-1 text-xs mr-1"
                          >
                            非受管
                          </div>
                        </span>
                      </p>
                    </div>
                    <div class="md:hidden flex space-x-1 truncate">
                      <span class="text-sm">{{
                        nn.Statics.latency != -1 ? nn.Statics.latency + "ms" : "断开"
                      }}</span
                      ><span>·</span
                      ><span class="md:hidden text-gray-600 text-sm">{{
                        nn.NoDERP ? "无中继" : "中继" + nn.DERPPort
                      }}</span
                      ><span>·</span
                      ><span class="md:hidden text-gray-600 text-sm">{{
                        nn.NoSTUN ? "无导航" : "导航" + nn.STUNPort
                      }}</span>
                    </div>
                    <div class="flex items-center text-gray-600 text-xs">
                      <span>{{ nn.Name }} </span>
                    </div>
                  </div>
                </td>
                <td class="hidden md:table-cell md:w-1/4">
                  <div class="flex relative min-w-0">
                    <div class="flex flex-col items-start text-gray-600 text-sm">
                      <span>IPv4: {{ nn.IPv4 == "" ? "未指定" : nn.IPv4 }} </span>
                      <span>IPv6: {{ nn.IPv6 == "" ? "未指定" : nn.IPv6 }} </span>
                    </div>
                  </div>
                </td>
                <td class="hidden md:table-cell w-1/4 lg:w-1/5">
                  <div class="flex relative min-w-0">
                    <div class="flex flex-col items-start text-sm">
                      <span>中继: {{ nn.NoDERP ? "已禁用" : nn.DERPPort }}</span>
                      <span>导航: {{ nn.NoSTUN ? "已禁用" : nn.STUNPort }}</span>
                    </div>
                  </div>
                </td>
                <td class="hidden lg:table-cell md:flex-auto">
                  <span>
                    <div class="inline-flex items-center cursor-default">
                      <span
                        class="inline-block w-2 h-2 rounded-full mr-2"
                        :class="{
                          'bg-green-500': nn.Statics.latency != -1,
                          'bg-gray-300': nn.Statics.latency == -1,
                        }"
                      ></span>
                      <span
                        v-if="nn.Arch != 'external'"
                        class="text-sm text-gray-600 tooltip tooltip-top"
                        :data-tip="
                          '已启动' + secondsFormat(nn.Statics.counter_uptime_sec)
                        "
                        >{{
                          nn.Statics.latency != -1 ? nn.Statics.latency + "ms" : "断开"
                        }}</span
                      >
                      <span
                        v-else
                        class="text-sm text-gray-600 tooltip tooltip-top"
                        data-tip="非受管司南"
                      >
                        {{
                          nn.Statics.latency != -1 ? nn.Statics.latency + "ms" : "断开"
                        }}
                      </span>
                    </div>
                  </span>
                </td>
                <td class="table-cell justify-end ml-auto md:ml-0 relative w-1/6 lg:w-12">
                  <div
                    v-if="
                      (!NaviNodeBtnShow && !NaviNodeMenuShow) ||
                      selectNaviNode.Name != nn.Name
                    "
                    @click="openNaviNodeMenu(nr, nn, $event)"
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

                  <div
                    v-if="
                      (NaviNodeBtnShow || NaviNodeMenuShow) &&
                      selectNaviNode.Name == nn.Name
                    "
                    @click="openNaviNodeMenu(nr, nn, $event)"
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
      </template>
    </section>
  </main>

  <!-- 提示框显示 -->
  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>

  <!--设备配置菜单显示-->
  <Teleport to="body">
    <NaviNodeMenu
      v-if="NaviNodeMenuShow"
      :toleft="btnLeft"
      :totop="btnTop"
      :select-navi="selectNaviNode"
      @close="closeNaviNodeMenu"
      @showdialog-removenavi="showRemoveNavi"
      @showdialog-detailinfo="showNaviDetails"
    ></NaviNodeMenu>
  </Teleport>

  <!-- 菜单弹出提示框显示 -->
  <Teleport to="body">
    <!--部署新司南提示框显示-->
    <Deploy
      v-if="deployDERPShow"
      :navi-region-list="NaviRegionList"
      @close="deployDERPShow = false"
      @add-done="addNaviDone"
    ></Deploy>

    <!-- 移除租户提示框显示 -->
    <RemoveNavi
      v-if="removeNaviShow"
      :select-navi="selectNaviNode"
      @close="removeNaviShow = false"
      @confirm-remove="doRemoveNavi"
    >
    </RemoveNavi>

    <!-- 编辑租户提示框显示 -->
    <NaviDetails
      v-if="naviDetailsShow"
      :select-navi="selectNaviNode"
      @close="naviDetailsShow = false"
    >
    </NaviDetails>
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
