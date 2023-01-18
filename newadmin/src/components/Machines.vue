<script setup>
import { ref, computed, nextTick, onMounted } from "vue";

//与框架交互部分

//界面控制部分
const toastShow = ref(false)
const toastMsg = ref("")

const delConfirmShow = ref(false)
const delMID = ref("1")

//数据填充控制部分
const MList = ref({});
const machinenumber = computed(() => {
    return Object.getOwnPropertyNames(MList.value).length;
});
onMounted(() => {
    axios
        .get("/admin/api/machines")
        .then(function (response) {
            // 处理成功情况
            if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
                for (var k in response.data["mlist"]) {
                    MList.value[k] = response.data["mlist"][k]
                    MList.value[k]["mipShow"] = false
                    MList.value[k]["menuShow"] = false
                    MList.value[k]["menuBtnShow"] = false
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
function removeMachine(id) {
    axios.
        post('/admin/api/machine/remove', {
            mid: id
        })
        .then(function (response) {
            if (response.data["status"] == "OK") {
                delConfirmShow.value = false
                MList.value[id] = nil
                rmRow = document.getElementById(id)
                rmRow.parentNode.removeChild(rmRow)
                alert("成功删除！")
            } else {
                alert("失败：" + response.data["errmsg"])
            }
        })
        .catch(function (error) {
            console.log(error);
        });
}

//客户端操作动作部分
function copyMIPv4(text) {
    navigator.clipboard.writeText(text).then(function () {
        toastMsg.value = "蜃境网络IPv4地址已复制到粘贴板！"
        toastShow.value = true
    });
}
function copyMIPv6(text) {
    navigator.clipboard.writeText(text).then(function () {
        toastMsg.value = "蜃境网络IPv6地址已复制到粘贴板！"
        toastShow.value = true
    });
}
function openOptionMenu(mID) {
    MList.value[mID]["menuShow"] = true
    document.body.style.pointerEvents = "none"
}
function closeOptionMenu(mID) {
    MList.value[mID]["menuShow"] = false
    document.body.style.removeProperty("pointer-events")
}

function showDelConfirm(id) {
    closeOptionMenu(id)
    MList.value[id]["menuBtnShow"] = false
    delMID.value = id
    delConfirmShow.value = true
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
                        <th class="md:w-1/4 flex-auto md:flex-initial md:shrink-0 w-0 text-ellipsis">设备</th>
                        <th class="hidden md:table-cell md:w-1/4">IP</th>
                        <th class="hidden md:table-cell w-1/4 lg:w-1/5">系统</th>
                        <th class="hidden lg:table-cell md:flex-auto">状态</th>
                        <th class="table-cell justify-end ml-auto md:ml-0 relative w-1/6 lg:w-12"><span
                                class="sr-only">设备操作菜单</span></th>
                    </tr>
                </thead>
                <tbody>
                    <template v-for="(m, id) in MList">
                        <tr :id="id" :v-if="MList[id] != nil" @mouseenter="m.menuBtnShow = true"
                            @mouseleave="m.menuBtnShow = false" class="w-full px-0.5 hover">
                            <td class="md:w-1/4 flex-auto md:flex-initial md:shrink-0 w-0 text-ellipsis">
                                <router-link class="relative" :to="'/machines/' + m.mipv4">
                                    <div class="items-center text-gray-900">
                                        <p class="font-semibold hover:text-blue-500">
                                            <span :class="{
                                                'bg-green-500': m.ifonline,
                                                'bg-gray-300': !m.ifonline,
                                            }"
                                                class="inline-block w-2 h-2 rounded-full relative -top-px lg:hidden mr-2"></span>
                                            <a class="stretched-link">{{
                                                m.givename
                                            }}
                                            </a>
                                        </p>
                                        <div class="md:hidden flex space-x-1 truncate"><span class="text-sm">{{
                                            m.mipv4
                                        }}</span><span>·</span><span class="md:hidden text-gray-600 text-sm"
                                                title="m.version">{{
                                                    m.os
                                                }}</span></div>
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
                                        <span v-if="m.issharedin">
                                            <div
                                                class="inline-flex items-center align-middle justify-center font-medium border border-orange-50 bg-orange-50 text-orange-600 rounded-sm px-1 text-xs mr-1">
                                                对外共享+1
                                            </div>
                                        </span>
                                        <span v-if="m.isexpirydisabled">
                                            <div
                                                class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs mr-1">
                                                永不过期
                                            </div>
                                        </span>
                                        <span v-if="m.isexitnode">
                                            <div
                                                class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                                                子网转发
                                            </div>
                                        </span>
                                        <span v-if="m.issubnet">
                                            <div
                                                class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                                                出口节点
                                            </div>
                                        </span>
                                    </div>
                                </div>
                            </td>
                            <td class="hidden md:table-cell md:w-1/4">
                                <ul>
                                    <li class="font-medium pr-6">
                                        <div @mouseenter="m.mipShow = true" @mouseleave="m.mipShow = false"
                                            class="flex relative min-w-0">
                                            <div class="truncate">
                                                <span>{{ m.mipv4 }} </span>
                                            </div>
                                            <div v-if="m.mipShow"
                                                class="absolute -mt-1 -ml-2 -top-px -left-px shadow-md cursor-pointer rounded-md active:shadow-sm transition-shadow duration-100 ease-in-out z-50"
                                                style="visibility: visible; max-width: 934px">
                                                <div class="flex border rounded-md button-outline bg-white">
                                                    <div @click="copyMIPv4(m.mipv4)"
                                                        class="flex min-w-0 py-1 px-2 hover:bg-gray-100 rounded-l-md">
                                                        <span class="inline-block select-none truncate"><span>
                                                                {{ m.mipv4 }}
                                                            </span></span><span
                                                            class="cursor-pointer text-blue-500 pl-2">复制</span>
                                                    </div>
                                                    <div @click="copyMIPv6(m.mipv6)"
                                                        class="text-blue-500 py-1 px-2 border-l hover:bg-gray-100 rounded-r-md">
                                                        IPv6
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </li>
                                    <template v-for="(mSub, mSubid) in m.msubnetlist">
                                        <li>
                                            <span>{{ mSub }} </span>
                                        </li>
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
                                        <span v-else class="text-sm text-gray-600 tooltip tooltip-top"
                                            :data-tip="'最近在线于' + m.lastseen">{{ m.lastseen }}
                                        </span>
                                    </div>
                                </span>
                            </td>
                            <td
                                class="table-cell justify-end ml-auto md:ml-0 relative w-12 justify-items-end items-center md:items-start">
                                <div v-if="!m.menuBtnShow && !m.menuShow"  @click="openOptionMenu(id)" class="flex-none w-12 -mt-0.5 relative">
                                    <button class="py-0.5 px-2 shadow-none rounded-md border border-gray-300/0
          group-hover:border-gray-300/100 hover:border-gray-300/100 group-hover:bg-white hover:!bg-gray-0
          group-hover:shadow-md hover:shadow-md hover:cursor-pointer active:border-gray-300/100 active:shadow focus:outline-none focus:ring transition-shadow
          duration-100 ease-in-out z-50"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" class="text-gray-500">
                                            <circle cx="12" cy="12" r="1"></circle>
                                            <circle cx="19" cy="12" r="1"></circle>
                                            <circle cx="5" cy="12" r="1"></circle>
                                        </svg>
                                    </button>
                                </div>
                                <!---->
                                <div v-if="m.menuBtnShow || m.menuShow" @click.self="openOptionMenu(id)"
                                    @blur="closeOptionMenu(id)" tabindex="-1" class="border button-outline bg-white shadow-md cursor-pointer divide-x divide-gray-200  active:shadow focus:outline-none focus:ring -mt-0.5 relative dropdown  dropdown-end py-0.5 px-2 rounded-md border-gray-300/0
          group-hover:border-gray-300/100 hover:border-gray-300/100 group-hover:bg-white hover:!bg-gray-0
          group-hover:shadow-md hover:shadow-md hover:cursor-pointer active:border-gray-300/100 transition-shadow
          duration-100 ease-in-out z-50 !border-y-0">

                                    <svg @click="openOptionMenu(id)" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="text-gray-500">
                                        <circle cx="12" cy="12" r="1"></circle>
                                        <circle cx="19" cy="12" r="1"></circle>
                                        <circle cx="5" cy="12" r="1"></circle>
                                    </svg>
                                    <div v-if="m.menuShow" 
                                        class="dropdown-content menu p-2 shadow bg-base-100 rounded-md w-52 px-0">
                                        <div class=" bg-white py-1 z-50"
                                            style="outline: none; --radix-dropdown-menu-content-transform-origin: var(--radix-popper-transform-origin); pointer-events: auto;">
                                            <div
                                                class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                                                编辑机器名称…</div>
                                            <div
                                                class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                                                分享…
                                            </div>
                                            <div
                                                class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                                                启用密钥过期
                                            </div>
                                            <div class="my-1 border-b border-gray-200"></div>
                                            <div
                                                class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                                                编辑子网…
                                            </div>
                                            <div
                                                class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                                                编辑标签…
                                            </div>
                                            <div class="my-1 border-b border-gray-200"></div>
                                            <div @click="showDelConfirm(id)"
                                                class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100 text-red-400">
                                                移除…
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </section>
    </main>

    <div v-if="toastShow" class="toast">
        <div class="alert shadow-lg bg-neutral text-neutral-content">
            <span>{{ toastMsg }}</span>
            <svg @click="toastShow = false" cursor="pointer" xmlns="http://www.w3.org/2000/svg"
                class="h-6 w-6 justify-self-end" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </div>
    </div>


    <div v-if="delConfirmShow" @click.self="delConfirmShow = false"
        class="fixed overflow-y-auto inset-0 py-8 z-10 bg-gray-900 bg-opacity-[0.07]" style="pointer-events: auto;">
        <div class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-dialog"
            tabindex="-1" style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">删除 {{ MList[delMID].givename }} </div>
            </header>
            <form @submit.prevent="removeMachine(delMID)">
                <p class="text-gray-700 mb-4">这个设备将从您的蜃境网络中永久删除！如需重新添加该设备，您将需要在该设备上重新进行授权</p>
                <footer class="flex mt-10 justify-end space-x-4">
                    <button class="btn bg-base-200 hover:bg-base-300 text-black hover:text-black border-0" type="button"
                        @click="delConfirmShow = false">取消</button>
                    <button class="btn bg-red-600 hover:bg-red-700 text-white border-0" type="submit">删除设备</button>
                </footer>
            </form>
            <button
                class="btn btn-square btn-ghost hover:bg-base-300 btn-sm absolute top-5 right-5 px-2 py-2 focus:bg-gray-100"
                type="button" @click="delConfirmShow = false">
                <svg xmlns="http://www.w3.org/2000/svg" width="1.25em" height="1.25em" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg></button>
        </div>
    </div>
</template>

<style scoped>

</style>
