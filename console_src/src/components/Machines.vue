<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch, watchEffect } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate } from "vue-router";
import MachineMenu from "./MachineMenu.vue";
import RemoveMachine from "./mmenu/RemoveMachine.vue";
import UpdateHostname from "./mmenu/UpdateHostname.vue";
import SetSubnet from "./mmenu/SetSubnet.vue";
import Toast from "./Toast.vue";

//与框架交互部分

//界面控制部分
const activeBtn = ref(null)
const btnLeft = ref(0)
const btnTop = ref(0)
function refreshMachineMenuPos() {
  if (activeBtn.value != null) {
    btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14
    btnTop.value = activeBtn.value?.getBoundingClientRect().top
  }
}
function openMachineMenu(mID, event) {
  activeBtn.value = event.target
  while (activeBtn.value?.tagName != "DIV" && activeBtn.value?.tagName != "div") {
    activeBtn.value = activeBtn.value.parentNode
  }
  currentMID.value = mID
  btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14
  btnTop.value = activeBtn.value?.getBoundingClientRect().top
  machineMenuShow.value = true;
}
function closeMachineMenu() {
  activeBtn.value = null
  machineMenuShow.value = false;
}

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () { toastShow.value = false }, 5000)
  }
})

const currentMID = ref("-1");
function mouseOnMachine(mid) {
  currentMID.value = mid
  machineBtnShow.value = true
}
function mouseLeaveMachine() {
  machineBtnShow.value = false
}
const machineIPShow = ref(false);
const machineMenuShow = ref(false);
const machineBtnShow = ref(false);

const delConfirmShow = ref(false);
function showDelConfirm() {
  machineBtnShow.value = false;
  closeMachineMenu(currentMID.value);
  delConfirmShow.value = true;
}
const updateHostnameShow = ref(false);
function showUpdateHostname() {
  machineBtnShow.value = false;
  closeMachineMenu(currentMID.value);
  updateHostnameShow.value = true;
}
const setSubnetShow = ref(false);
function showSetSubnet() {
  machineBtnShow.value = false;
  closeMachineMenu(currentMID.value);
  setSubnetShow.value = true;
}

//数据填充控制部分
const MList = ref({});
const machinenumber = computed(() => {
  return Object.getOwnPropertyNames(MList.value).length;
});
let getMIntID;
function getMachines() {
  return new Promise((resolve, reject) => {
    axios
      .get("/admin/api/machines")
      .then(function (response) {
        if (response.data["needreauth"] != undefined || response.data["needreauth"] == true) {
          toastMsg.value = response.data["needreauthreason"] + "，登录状态失效，请重新登录";
          toastShow.value = true;
          reject()
        }
        // 处理成功情况
        if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
          for (var k in response.data["mlist"]) {
            MList.value[k] = response.data["mlist"][k];
            let tailtwo = MList.value[k]["expirydesc"].slice(-2);
            let tailthree = MList.value[k]["expirydesc"].slice(-3);
            if (
              MList.value[k]["expirydesc"] == "马上就要过期" ||
              tailtwo == "分钟" ||
              tailtwo == "小时" ||
              tailthree == "剩1天"
            ) {
              MList.value[k]["soonexpiry"] = true;
            } else {
              MList.value[k]["soonexpiry"] = false;
            }
          }
          resolve();
        } else if (response.data["errormsg"] != undefined) {
          toastMsg.value = "获设备信息出错：" + response.data["errormsg"];
          toastShow.value = true;
          reject();
        }
      })
      .catch(function (error) {
        // 处理错误情况
        toastMsg.value = "更新页面出错：" + error;
        toastShow.value = true;
        reject();
      })
      .then(function () {
        // 总是会执行
      });
  });
}
onMounted(() => {
  refreshMachineMenuPos()
  window.addEventListener("resize", refreshMachineMenuPos)
  window.addEventListener("scroll", refreshMachineMenuPos)


  getMachines().then().catch();
  getMIntID = setInterval(() => {
    getMachines().then().catch();
  }, 15000);
});
onUnmounted(() => {
  window.removeEventListener("resize", refreshMachineMenuPos)
  window.removeEventListener("scroll", refreshMachineMenuPos)
})
onBeforeRouteLeave(() => {
  clearInterval(getMIntID);
});

//服务端请求
function setExpires(id) {
  closeMachineMenu()
  axios
    .post("/admin/api/machines", {
      mid: id,
      state: "set-expires"
    })
    .then(function (response) {
      if (response.data["status"] == "success") {
        MList.value[id]["neverExpires"] = response.data["data"]["neverExpires"]
        MList.value[id]["expirydesc"] = response.data["data"]["expires"]
        if (response.data["data"]["neverExpires"] == true) {
          toastMsg.value = "已禁用密钥过期";
        } else {
          toastMsg.value = "已启用密钥过期";
        }
        toastShow.value = true;
      } else {
        toastMsg.value = "失败：" + response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      console.log(error);
    });
}
function removeMachine(id) {
  axios
    .post("/admin/api/machine/remove", {
      mid: id,
    })
    .then(function (response) {
      if (response.data["status"] == "OK") {
        delConfirmShow.value = false;
        toastMsg.value = MList.value[id]["name"] + "已从您的蜃境网络移除！";
        toastShow.value = true;
        delete MList.value[id];
      } else {
        alert("失败：" + response.data["errmsg"]);
      }
    })
    .catch(function (error) {
      console.log(error);
    });
}

function hostnameUpdateDone(newName, newAutomaticNameMode, wantClose) {
  MList.value[currentMID.value]["name"] = newName
  MList.value[currentMID.value]["automaticNameMode"] = newAutomaticNameMode
  nextTick(() => {
    updateHostnameShow.value = !wantClose
    nextTick(() => {
      toastMsg.value = "已更新设备名称！"
      toastShow.value = true
    })
  })
}
function hostnameUpdateFail(msg) {
  toastMsg.value = "更新设备名称失败！"
  toastShow.value = true
}

function subnetUpdateDone(newAllIPs, newAllowedIPs, newExtraIPs, newEnExitNode) {
  MList.value[currentMID.value]["advertisedIPs"] = newAllIPs
  MList.value[currentMID.value]["allowedIPs"] = newAllowedIPs
  MList.value[currentMID.value]["extraIPs"] = newExtraIPs
  MList.value[currentMID.value]["allowedExitNode"] = newEnExitNode
  nextTick(() => {
    toastMsg.value = "已更新子网转发设置！"
    toastShow.value = true
  })
}
function subnetUpdateFail(msg) {
  toastMsg.value = "更新子网转发设置失败！"
  toastShow.value = true
}

//客户端操作动作部分
function copyMIPv4() {
  navigator.clipboard.writeText(MList.value[currentMID.value]["mipv4"]).then(function () {
    toastMsg.value = "蜃境网络IPv4地址已复制到粘贴板！";
    toastShow.value = true;
  });
}
function copyMIPv6() {
  navigator.clipboard.writeText(MList.value[currentMID.value]["mipv6"]).then(function () {
    toastMsg.value = "蜃境网络IPv6地址已复制到粘贴板！";
    toastShow.value = true;
  });
}
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="mb-24">
      <header class="mb-8">
        <div class="flex justify-between items-center">
          <div class="flex items-center">
            <h1 class="text-3xl font-semibold tracking-tight leading-tight mb-2" tabindex="-1">
              设备
            </h1>
          </div>
        </div>
      </header>

      <div
        class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-full px-2 py-1 leading-none text-sm mb-8">
        {{ machinenumber }} 个设备
      </div>
      <table class="table w-full">
        <thead>
          <tr>
            <th class="md:w-1/4 flex-auto md:flex-initial md:shrink-0 w-0 text-ellipsis">
              设备
            </th>
            <th class="hidden md:table-cell md:w-1/4">IP</th>
            <th class="hidden md:table-cell w-1/4 lg:w-1/5">系统</th>
            <th class="hidden lg:table-cell md:flex-auto">状态</th>
            <th class="table-cell justify-end ml-auto md:ml-0 relative w-1/6 lg:w-12">
              <span class="sr-only">设备操作菜单</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(m, id) in MList">
            <tr :id="id" :v-if="MList[id] != nil" @mouseenter="mouseOnMachine(id)" @mouseleave="mouseLeaveMachine(id)"
              class="w-full px-0.5 hover">
              <td class="md:w-1/4 flex-auto md:flex-initial md:shrink-0 w-0 text-ellipsis">
                <router-link class="relative" :to="'/machines/' + m.mipv4">
                  <div class="items-center text-gray-900">
                    <p class="font-semibold hover:text-blue-500">
                      <span :class="{
                        'bg-green-500': m.ifonline,
                        'bg-gray-300': !m.ifonline,
                      }" class="inline-block w-2 h-2 rounded-full relative -top-px lg:hidden mr-2"></span>
                      <a class="stretched-link">{{ m.name }} </a>
                    </p>
                    <div class="md:hidden flex space-x-1 truncate">
                      <span class="text-sm">{{ m.mipv4 }}</span><span>·</span><span
                        class="md:hidden text-gray-600 text-sm" title="m.version">{{
                          m.os
                        }}</span>
                    </div>
                  </div>
                  <div>
                    <div class="flex items-center text-gray-600 text-sm">
                      <span>{{ m.useraccount }} </span>
                    </div>
                  </div>
                </router-link>
                <div class="my-1">
                  <div>
                    <span v-if="m.issharedin">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-orange-50 bg-orange-50 text-orange-600 rounded-sm px-1 text-xs mr-1">
                        外部共享
                      </div>
                    </span>
                    <span v-if="m.issharedout">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-orange-50 bg-orange-50 text-orange-600 rounded-sm px-1 text-xs mr-1">
                        对外共享+1
                      </div>
                    </span>
                    <span v-if="m.expirydesc == '已过期'">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-red-50 bg-red-50 text-red-600 rounded-sm px-1 text-xs mr-1">
                        已过期
                      </div>
                    </span>
                    <span v-if="m.neverExpires">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs mr-1">
                        永不过期
                      </div>
                    </span>
                    <span v-if="m.soonexpiry">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs mr-1">
                        {{ m.expirydesc }}
                      </div>
                    </span>
                    <span v-if="m.hasSubnets">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                        子网转发
                        <div v-if="m.hasSubnets && m.extraIPs && m.extraIPs.length > 0" class="tooltip"
                          data-tip="该设备存在未批准子网转发，请在设备菜单的“编辑子网转发…”中检查">
                          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor" stroke-width="2.35" stroke-linecap="round"
                            stroke-linejoin="round" class="ml-1">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="8" x2="12" y2="12"></line>
                            <line x1="12" y1="16" x2="12.01" y2="16"></line>
                          </svg>
                        </div>
                      </div>
                    </span>
                    <span v-if="m.advertisedExitNode">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                        出口节点
                        <div v-if="!m.allowedExitNode" class="tooltip" data-tip="该设备申请被用作出口节点，请在设备菜单的“编辑子网转发…”中检查">
                          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor" stroke-width="2.35" stroke-linecap="round"
                            stroke-linejoin="round" class="ml-1">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="8" x2="12" y2="12"></line>
                            <line x1="12" y1="16" x2="12.01" y2="16"></line>
                          </svg>
                        </div>
                      </div>
                    </span>
                  </div>
                </div>
              </td>
              <td class="hidden md:table-cell md:w-1/4">
                <ul>
                  <li class="font-medium pr-6">
                    <div @mouseenter="machineIPShow = true" @mouseleave="machineIPShow = false"
                      class="flex relative min-w-0">
                      <div class="truncate">
                        <span>{{ m.mipv4 }} </span>
                      </div>
                      <div v-if="machineIPShow && currentMID == id"
                        class="absolute -mt-1 -ml-2 -top-px -left-px shadow-md cursor-pointer rounded-md active:shadow-sm transition-shadow duration-100 ease-in-out z-20"
                        style="visibility: visible; max-width: 934px">
                        <div class="flex border rounded-md button-outline bg-white">
                          <div @click="copyMIPv4" class="flex min-w-0 py-1 px-2 hover:bg-gray-100 rounded-l-md">
                            <span class="inline-block select-none truncate"><span>
                                {{ m.mipv4 }}
                              </span></span><span class="cursor-pointer text-blue-500 pl-2">复制</span>
                          </div>
                          <div @click="copyMIPv6"
                            class="text-blue-500 py-1 px-2 border-l hover:bg-gray-100 rounded-r-md">
                            IPv6
                          </div>
                        </div>
                      </div>
                    </div>
                  </li>
                  <li v-for="allowedIP in m.allowedIPs">
                    <span>{{ allowedIP }} </span>
                  </li>
                  <template v-for="extraIP in m.extraIPs">
                    <li class="tooltip text-gray-400" data-tip="这条子网转发未启用">
                      <span>{{ extraIP }} </span>
                    </li><br />
                  </template>
                </ul>
              </td>
              <td class="hidden md:table-cell w-1/4 lg:w-1/5">
                <div class="flex items-center relative">
                  <div>{{ m.os }}</div>
                </div>
                <div class="text-sm text-gray-600">{{ m.version }}</div>
              </td>
              <td class="hidden lg:table-cell md:flex-auto">
                <span>
                  <div class="inline-flex items-center cursor-default">
                    <span class="inline-block w-2 h-2 rounded-full mr-2" :class="{
                      'bg-green-500': m.ifonline,
                      'bg-gray-300': !m.ifonline,
                    }"></span>
                    <span v-if="m.ifonline" class="text-sm text-gray-600 tooltip tooltip-top"
                      :data-tip="'最近在线于' + m.lastseen">已连接</span>
                    <span v-else class="text-sm text-gray-600 tooltip tooltip-top" :data-tip="'最近在线于' + m.lastseen">{{
                      m.lastseen
                    }}
                    </span>
                  </div>
                </span>
              </td>
              <td
                class="table-cell justify-end ml-auto md:ml-0 relative w-12 justify-items-end items-center md:items-start">
                <div v-if="!machineBtnShow && !machineMenuShow || currentMID != id" @click="openMachineMenu(id, $event)"
                  class="flex-none w-12 -mt-0.5 relative">
                  <button
                    class="py-0.5 px-2 shadow-none rounded-md border border-gray-300/0 hover:border-gray-300/100 hover:bg-gray-100 hover:shadow-md hover:cursor-pointer active:border-gray-300/100 active:shadow focus:outline-none focus:ring transition-shadow duration-100 ease-in-out z-20">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="text-gray-500">
                      <circle cx="12" cy="12" r="1"></circle>
                      <circle cx="19" cy="12" r="1"></circle>
                      <circle cx="5" cy="12" r="1"></circle>
                    </svg>
                  </button>
                </div>
                <!---->
                <div v-if="(machineBtnShow || machineMenuShow) && currentMID == id" @click="openMachineMenu(id, $event)"
                  class="flex-none w-12 border button-outline bg-white shadow-md cursor-pointer focus:outline-none focus:ring -mt-0.5 relative py-0.5 px-2 rounded-md border-gray-300/100 hover:border-gray-300/100 hover:bg-gray-100 hover:shadow-md hover:cursor-pointer active:border-gray-300/100 transition-shadow duration-100 ease-in-out z-20 ">
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="text-gray-500">
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
    <MachineMenu v-if="machineMenuShow" :toleft="btnLeft" :totop="btnTop" :neverExpires="MList[currentMID].neverExpires"
      @close="closeMachineMenu" @set-expires="setExpires(currentMID)" @showdialog-remove="showDelConfirm"
      @showdialog-updatehostname="showUpdateHostname" @showdialog-setsubnet="showSetSubnet"></MachineMenu>
  </Teleport>

  <!-- 菜单弹出提示框显示 -->
  <Teleport to="body">
    <!-- 删除设备提示框显示 -->
    <RemoveMachine v-if="delConfirmShow" :machine-name="MList[currentMID].name" @close="delConfirmShow = false"
      @confirm="removeMachine(currentMID)"></RemoveMachine>
    <!-- 修改设备名提示框显示 -->
    <UpdateHostname v-if="updateHostnameShow" :id="currentMID" :given-name="MList[currentMID].name"
      :host-name="MList[currentMID].hostname" :auto-gen="MList[currentMID].automaticNameMode"
      @close="updateHostnameShow = false" @update-done="hostnameUpdateDone" @update-fail="hostnameUpdateFail">
    </UpdateHostname>
    <!-- 设置子网转发提示框显示 -->
    <SetSubnet v-if="setSubnetShow" :id="currentMID" :current-machine="MList[currentMID]" @close="setSubnetShow = false"
      @update-done="subnetUpdateDone" @update-fail="subnetUpdateFail"></SetSubnet>
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
