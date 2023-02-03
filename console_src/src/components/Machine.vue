<script setup>
import { ref, computed, nextTick, watch, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import MachineMenu from "./MachineMenu.vue";
import RemoveMachine from "./mmenu/RemoveMachine.vue";
import UpdateHostname from "./mmenu/UpdateHostname.vue";
import SetSubnet from "./mmenu/SetSubnet.vue";
import Toast from "./Toast.vue";

const devmode = ref(false);

const router = useRouter();
const route = useRoute();

//界面控制部分
const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
    if (toastShow.value) {
        setTimeout(function () { toastShow.value = false }, 5000)
    }
})

const hasSpecialStatus = computed(() => {
    return (
        currentMachine.value["issharedin"] ||
        currentMachine.value["issharedout"] ||
        currentMachine.value["expirydesc"] == "已过期" ||
        currentMachine.value["neverExpires"] ||
        currentMachine.value["soonexpiry"] ||
        currentMachine.value["advertisedExitNode"] ||
        currentMachine.value["hasSubnets"]
    );
});

const activeBtn = ref(null)
const btnLeft = ref(0)
const btnTop = ref(0)
const machineMenuShow = ref(false)
function watchWindowChange() {
    if (activeBtn.value != null) {
        btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 78
        btnTop.value = activeBtn.value?.getBoundingClientRect().top + 20
    }
    window.onresize = () => {
        if (activeBtn.value != null) {
            btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 78
            btnTop.value = activeBtn.value?.getBoundingClientRect().top + 20
        }
    }
    window.onscroll = () => {
        if (activeBtn.value != null) {
            btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 78
            btnTop.value = activeBtn.value?.getBoundingClientRect().top + 20
        }
    }
}
function openMachineMenu(event) {
    activeBtn.value = event.target
    while (activeBtn.value?.tagName != "BUTTON" && activeBtn.value?.tagName != "button") {
        activeBtn.value = activeBtn.value?.parentNode
    }
    btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 78
    btnTop.value = activeBtn.value?.getBoundingClientRect().top + 20
    machineMenuShow.value = true;
}
function closeMachineMenu() {
    activeBtn.value = null
    machineMenuShow.value = false;
}

const delConfirmShow = ref(false);
function showDelConfirm() {
    closeMachineMenu();
    delConfirmShow.value = true;
}
const updateHostnameShow = ref(false);
function showUpdateHostname() {
    closeMachineMenu();
    updateHostnameShow.value = true;
}
const setSubnetShow = ref(false);
function showSetSubnet() {
    closeMachineMenu();
    setSubnetShow.value = true;
}

//数据填充控制部分
const currentMachine = ref({});
const currentMID = ref("");
const basedomain = ref("");
onMounted(() => {
    watchWindowChange()
    axios
        .get("/admin/api/machines")
        .then(function (response) {
            if (
                response.data["needreauth"] != undefined ||
                response.data["needreauth"] == true
            ) {
                toastMsg.value = response.data["needreauthreason"] + "，登录状态失效，请重新登录";
                toastShow.value = true;
                reject();
            }
            // 处理成功情况
            if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
                basedomain.value = response.data["basedomain"];
                for (var k in response.data["mlist"]) {
                    if (response.data["mlist"][k]["mipv4"] === route.params.mip) {
                        currentMachine.value = response.data["mlist"][k];
                        currentMID.value = k;
                        let tailtwo = currentMachine.value["expirydesc"].slice(-2);
                        if (
                            currentMachine.value["expirydesc"] == "马上就要过期" ||
                            tailtwo == "分钟" ||
                            tailtwo == "小时" ||
                            tailtwo == "1天"
                        ) {
                            currentMachine.value["soonexpiry"] = true;
                        } else {
                            currentMachine.value["soonexpiry"] = false;
                        }
                        break;
                    }
                }
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
//服务端请求
function setExpires() {
    closeMachineMenu()
    axios
        .post("/admin/api/machines", {
            mid: currentMID.value,
            state: "set-expires"
        })
        .then(function (response) {
            if (response.data["status"] == "success") {
                currentMachine.value["neverExpires"] = response.data["data"]["neverExpires"]
                currentMachine.value["expirydesc"] = response.data["data"]["expires"]
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
function removeMachine() {
    axios
        .post("/admin/api/machine/remove", {
            mid: currentMID,
        })
        .then(function (response) {
            if (response.data["status"] == "OK") { //TODO: 需处理设备页面删除跳转后的Toast显示
                delConfirmShow.value = false;
                toastMsg.value = currentMachine.value["name"] + "已从您的蜃境网络移除！";
                toastShow.value = true;

                router.push('/machines')
            } else {
                alert("失败：" + response.data["errmsg"]);
            }
        })
        .catch(function (error) {
            console.log(error);
        });
}
function hostnameUpdateDone(newName, newAutomaticNameMode, wantClose) {
    currentMachine.value["name"] = newName
    currentMachine.value["automaticNameMode"] = newAutomaticNameMode
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
    currentMachine.value["advertisedIPs"] = newAllIPs
    currentMachine.value["allowedIPs"] = newAllowedIPs
    currentMachine.value["extraIPs"] = newExtraIPs
    currentMachine.value["allowedExitNode"] = newEnExitNode
    nextTick(() => {
        toastMsg.value = "已更新子网转发设置！"
        toastShow.value = true
    })
}
function subnetUpdateFail(msg) {
    toastMsg.value = "更新子网转发设置失败！"
    toastShow.value = true
}

</script>

<template>
    <main class="container mx-auto pb-20 md:pb-24">
        <section class="mb-24">
            <header class="pb-4 mb-8">
                <div class="font-medium space-x-2 mb-5 truncate flex">
                    <router-link to="/machines" class="text-blue-500">全部设备</router-link><span
                        class="text-gray-400">/</span><span>{{ currentMachine.mipv4 }}</span>
                </div>
                <div class="flex flex-wrap gap-2 items-center justify-between">
                    <h1 class="text-2xl font-semibold tracking-tight leading-tight truncate flex-shrink-0 max-w-full"
                        tabindex="-1">
                        {{ currentMachine.name }}
                    </h1>
                    <div class="flex">
                        <div class="flex gap-2 flex-wrap">
                            <button @click="openMachineMenu($event)"
                                class="btn btn-outline bg-gray-50 border-gray-300 text-gray-700 hover:bg-gray-100 hover:border-gray-300 hover:text-gray-700 min-w-0">
                                <div class="flex items-center">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="mr-2">
                                        <circle cx="12" cy="12" r="3"></circle>
                                        <path
                                            d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z">
                                        </path>
                                    </svg>设备设置
                                </div>
                            </button>
                        </div>
                    </div>
                </div>
                <div class="flex border-t border-gray-200 text-sm mt-4 pt-4">
                    <div class="max-w-sm">
                        <div class="text-gray-500 mb-2">归属于</div>
                        <div class="mt-0.5">
                            <div class="flex items-center text-gray-600 text-sm">
                                <div class="relative shrink-0 rounded-full overflow-hidden w-5 h-5 text-xs mr-2">
                                    <div class="flex items-center justify-center text-center capitalize text-white font-medium pointer-events-none w-5 h-5 text-xs"
                                        style="background-color: rgb(161, 56, 33)">
                                        {{ currentMachine.usernamehead }}
                                    </div>
                                </div>
                                <span data-state="closed">{{ currentMachine.useraccount }}</span>
                            </div>
                        </div>
                    </div>
                    <div v-if="hasSpecialStatus" class="max-w-sm border-l border-gray-200 ml-4 pl-4">
                        <p class="text-gray-500 mb-2">状态</p>
                        <div>
                            <span v-if="currentMachine.issharedin">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-orange-50 bg-orange-50 text-orange-600 rounded-sm px-1 text-xs mr-1">
                                    外部共享
                                </div>
                            </span>
                            <span v-if="currentMachine.issharedout">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-orange-50 bg-orange-50 text-orange-600 rounded-sm px-1 text-xs mr-1">
                                    对外共享+1
                                </div>
                            </span>
                            <span v-if="currentMachine.expirydesc == '已过期'">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-red-50 bg-red-50 text-red-600 rounded-sm px-1 text-xs mr-1">
                                    已过期
                                </div>
                            </span>
                            <span v-if="currentMachine.neverExpires">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs mr-1">
                                    永不过期
                                </div>
                            </span>

                            <span v-if="currentMachine.soonexpiry">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs mr-1">
                                    {{ currentMachine.expirydesc }}
                                </div>
                            </span>
                            <span v-if="currentMachine.hasSubnets">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                                    子网转发
                                    <div v-if="currentMachine.hasSubnets && currentMachine.extraIPs && currentMachine.extraIPs.length > 0"
                                        class="tooltip" data-tip="该设备存在未批准子网转发，请在设备菜单的“编辑子网转发…”中检查">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.35"
                                            stroke-linecap="round" stroke-linejoin="round" class="ml-1">
                                            <circle cx="12" cy="12" r="10"></circle>
                                            <line x1="12" y1="8" x2="12" y2="12"></line>
                                            <line x1="12" y1="16" x2="12.01" y2="16"></line>
                                        </svg>
                                    </div>
                                </div>
                            </span>
                            <span v-if="currentMachine.advertisedExitNode">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                                    出口节点
                                    <div v-if="!currentMachine.allowedExitNode" class="tooltip"
                                        data-tip="该设备申请被用作出口节点，请在设备菜单的“编辑子网转发…”中检查">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.35"
                                            stroke-linecap="round" stroke-linejoin="round" class="ml-1">
                                            <circle cx="12" cy="12" r="10"></circle>
                                            <line x1="12" y1="8" x2="12" y2="12"></line>
                                            <line x1="12" y1="16" x2="12.01" y2="16"></line>
                                        </svg>
                                    </div>
                                </div>
                            </span>
                        </div>
                    </div>
                </div>
            </header>
            <section class="mb-8">
                <header class="flex justify-between mb-4">
                    <div class="max-w-xl">
                        <h3 class="text-xl font-semibold tracking-tight mb-2">子网转发</h3>
                        <p class="text-gray-600">
                            “子网转发”允许你暴露设备可访问物理网络路由给您的蜃境网络
                        </p>
                    </div>
                    <div v-if="currentMachine.hasSubnets">
                        <button @click="showSetSubnet"
                            class="btn btn-outline bg-gray-50 border-gray-300 text-gray-700 hover:bg-gray-100 hover:border-gray-300 hover:text-gray-700 mt-2">
                            配置
                        </button>
                    </div>
                </header>
                <div v-if="currentMachine.hasSubnets" class="p-4 md:p-6 border border-gray-200 rounded-md">
                    <ul class="leading-normal">
                        <li v-for="allowedIP in currentMachine.allowedIPs">
                            <span>{{ allowedIP }} </span>
                        </li>
                        <template v-for="extraIP in currentMachine.extraIPs">
                            <li class="tooltip text-gray-400" data-tip="这条子网转发未启用">
                                <span>{{ extraIP }} </span>
                            </li><br />
                        </template>
                    </ul>
                </div>
                <div v-if="!currentMachine.hasSubnets && !currentMachine.issharedin"
                    class="p-4 md:p-6 border border-gray-200 rounded-md flex items-center justify-center text-gray-500 text-center">
                    <div class="flex justify-center">
                        <div class="w-full text-center max-w-xl text-gray-500">
                            该设备未暴露任何子网可供转发
                        </div>
                    </div>
                </div>
                <div v-if="currentMachine.issharedin"
                    class="p-4 md:p-6 border border-gray-200 rounded-md flex items-center justify-center text-gray-500 text-center">
                    <div class="flex justify-center">
                        <div class="w-full text-center max-w-xl text-gray-500">
                            该设备来自外部共享，不能暴露子网转发给你
                        </div>
                    </div>
                </div>
            </section>
            <section class="mb-8">
                <header class="max-w-xl mb-4">
                    <h3 class="text-xl font-semibold tracking-tight mb-2">设备信息</h3>
                    <p class="text-gray-600">关于该设备网络的信息，用于调试连接问题</p>
                </header>
                <div
                    class="p-4 md:p-6 border border-gray-200 rounded-md grid grid-cols-1 md:grid-cols-2 gap-y-2 gap-x-2">
                    <div class="space-y-2">
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">创建者</dt>
                            <dd class="min-w-0 truncate">{{ currentMachine.useraccount }}</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">设备名称</dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate">{{ currentMachine.name }}</div>
                                    <div v-if="devmode" class="cursor-pointer text-blue-500 pl-2">复制</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">域名</dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate">
                                        {{ currentMachine.name }}.{{ currentMachine.useraccount }}.{{
                                            basedomain
                                        }}
                                    </div>
                                    <div v-if="devmode" class="cursor-pointer text-blue-500 pl-2">复制</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">系统主机名</dt>
                            <dd class="min-w-0 truncate">{{ currentMachine.hostname }}</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">操作系统</dt>
                            <dd class="min-w-0 truncate">{{ currentMachine.os }}</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">蜃境客户端版本</dt>
                            <dd class="min-w-0 truncate">
                                <div class="flex items-center">{{ currentMachine.version }}</div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">蜃境网络 IPv4</dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate">
                                        <span>{{ currentMachine.mipv4 }}</span>
                                    </div>
                                    <div v-if="devmode" class="cursor-pointer text-blue-500 pl-2">复制</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">蜃境网络 IPv6</dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate">
                                        <span class="inline-flex justify-start min-w-0 max-w-full"><span
                                                class="truncate w-fit flex-shrink">{{
                                                    currentMachine.mipv6
                                                }}</span></span>
                                    </div>
                                    <div v-if="devmode" class="cursor-pointer text-blue-500 pl-2">复制</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">设备ID</dt>
                            <dd class="min-w-0 truncate">{{ currentMID }}</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">设备端点信息</dt>
                            <dd class="min-w-0 truncate">
                                <ul class="pl-3 -indent-3">
                                    <li v-for="ep in currentMachine.endpoints" class="select-all">
                                        <span>{{ ep }}</span>
                                    </li>
                                    <!--
                                    <li class="select-all">
                                        <span>172.17.0.1</span><wbr />:<span>41641</span>
                                    </li>
                                    -->
                                </ul>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">中继器</dt>
                            <dd class="min-w-0 truncate">
                                <ul v-if="currentMachine.usederp == 'x'">
                                    <li>
                                        中继器未设定或出错
                                    </li>
                                </ul>
                                <ul v-else>
                                    <li v-for="(latency, derpname) in currentMachine.derps">
                                        <strong class="font-medium">{{ derpname }} 号中继</strong>: {{ latency }}&nbsp;ms
                                        <svg v-if="currentMachine.usederp == derpname"
                                            xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="relative inline-block ml-1 -top-px">
                                            <polyline points="20 6 9 17 4 12"></polyline>
                                        </svg>
                                    </li>
                                </ul>
                            </dd>
                        </dl>
                    </div>
                    <div class="space-y-2">
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">创建于</dt>
                            <dd class="min-w-0 truncate">{{ currentMachine.createat }}</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">最近更新</dt>
                            <dd class="min-w-0 truncate">{{ currentMachine.lastseen }}</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">密钥过期</dt>
                            <dd v-if="currentMachine.neverExpires" class="min-w-0 truncate">
                                永不过期
                            </dd>
                            <dd v-if="!currentMachine.neverExpires" class="min-w-0 truncate">
                                {{ currentMachine.expirydesc }}
                            </dd>
                        </dl>
                        <h2 class="pt-2 text-xs uppercase font-semibold text-gray-500 tracking-wide">
                            客户端连通性
                        </h2>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">复杂网络 Varies</dt>
                            <dd v-if="currentMachine.varies" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">
                                需发夹机制 Hairpinning
                            </dt>
                            <dd v-if="currentMachine.hairpinning" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">IPv6</dt>
                            <dd v-if="currentMachine.ipv6en" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">UDP</dt>
                            <dd v-if="currentMachine.udpen" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">UPnP</dt>
                            <dd v-if="currentMachine.upnpen" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">PCP</dt>
                            <dd v-if="currentMachine.pcpen" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">NAT-PMP</dt>
                            <dd v-if="currentMachine.pmpen" class="min-w-0 truncate">是</dd>
                            <dd v-else class="min-w-0 truncate">否</dd>
                        </dl>
                    </div>
                </div>
            </section>
        </section>
    </main>

    <!-- 提示框显示 -->
    <Teleport to=".toast-container">
        <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
    </Teleport>

    <!--设备配置菜单显示-->
    <Teleport to="body">
        <MachineMenu v-if="machineMenuShow" :toleft="btnLeft" :totop="btnTop"
            :neverExpires="currentMachine.neverExpires" @close="closeMachineMenu" @set-expires="setExpires"
            @showdialog-remove="showDelConfirm" @showdialog-updatehostname="showUpdateHostname"
            @showdialog-setsubnet="showSetSubnet"></MachineMenu>
    </Teleport>

    <!-- 菜单弹出提示框显示 -->
    <Teleport to="body">
        <!-- 删除设备提示框显示 -->
        <RemoveMachine v-if="delConfirmShow" :machine-name="currentMachine.name" @close="delConfirmShow = false"
            @confirm="removeMachine"></RemoveMachine>
        <!-- 修改设备名提示框显示 -->
        <UpdateHostname v-if="updateHostnameShow" :id="currentMID" :host-name="currentMachine.hostname"
            :given-name="currentMachine.name" :auto-gen="currentMachine.automaticNameMode"
            @close="updateHostnameShow = false" @update-done="hostnameUpdateDone" @update-fail="hostnameUpdateFail">
        </UpdateHostname>
        <!-- 设置子网转发提示框显示 -->
        <SetSubnet v-if="setSubnetShow" :id="currentMID" :current-machine="currentMachine"
            @close="setSubnetShow = false" @update-done="subnetUpdateDone" @update-fail="subnetUpdateFail"></SetSubnet>
    </Teleport>
</template>

<style scoped>
.table tr.hover:hover th,
.table tr.hover:hover td,
.table tr.hover:nth-child(even):hover th,
.table tr.hover:nth-child(even):hover td {
    background-color: #faf9f8;
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
