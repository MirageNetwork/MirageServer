<script setup>
import { ref, computed, nextTick, onMounted } from "vue";
import { useRouter, useRoute } from 'vue-router';

const router = useRouter()
const route = useRoute()

//数据填充控制部分
const currentMachine = ref({});
onMounted(() => {
    axios
        .get("/admin/api/machines")
        .then(function (response) {
            // 处理成功情况
            if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
                for (var k in response.data["mlist"]) {
                    if (response.data["mlist"][k]["mipv4"] === route.params.mip) {
                        currentMachine.value = response.data["mlist"][k]
                        break
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
</script>

<template>
    <main class="container mx-auto pb-20 md:pb-24">
        <section class="mb-24">
            <header class="pb-4 mb-8">
                <div class="font-medium space-x-2 mb-5 truncate flex"><router-link to="/machines"
                        class="text-blue-500">全部设备</router-link><span class="text-gray-400">/</span><span>{{
                            currentMachine.mipv4
                        }}</span></div>
                <div class="flex flex-wrap gap-2 items-center justify-between">
                    <h1 class="text-2xl font-semibold tracking-tight leading-tight truncate flex-shrink-0 max-w-full"
                        tabindex="-1">{{ currentMachine.givename }}</h1>
                    <div class="flex">
                        <div class="flex gap-2 flex-wrap"><button class="button button-outline min-w-0" type="button"
                                id="radix-:r5n:" aria-haspopup="menu" aria-expanded="false" data-state="closed">
                                <div class="flex items-center"><svg xmlns="http://www.w3.org/2000/svg" width="16"
                                        height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-2">
                                        <circle cx="12" cy="12" r="3"></circle>
                                        <path
                                            d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z">
                                        </path>
                                    </svg>设备设置</div>
                            </button></div>
                    </div>
                </div>
                <div class="flex border-t border-gray-200 text-sm mt-4 pt-4">
                    <div class="max-w-sm">
                        <div class="text-gray-500 mb-2">归属于</div>
                        <div class="mt-0.5">
                            <div class="flex items-center text-gray-600 text-sm">
                                <div class="relative shrink-0 rounded-full overflow-hidden w-5 h-5 text-xs mr-2">
                                    <div class="flex items-center justify-center text-center capitalize text-white font-medium pointer-events-none w-5 h-5 text-xs"
                                        style="background-color: rgb(161, 56, 33);">{{ currentMachine.usernamehead }}
                                    </div>
                                </div><span data-state="closed">{{ currentMachine.useraccount }}</span>
                            </div>
                        </div>
                    </div>
                    <div class="max-w-sm border-l border-gray-200 ml-4 pl-4">
                        <p class="text-gray-500 mb-2">状态</p>
                        <div><span data-state="closed">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs mr-1">
                                    Expiry disabled</div>
                            </span><span data-state="closed">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                                    Subnets<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.35"
                                        stroke-linecap="round" stroke-linejoin="round" class="ml-1">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="8" x2="12" y2="12"></line>
                                        <line x1="12" y1="16" x2="12.01" y2="16"></line>
                                    </svg></div>
                            </span><span data-state="closed">
                                <div
                                    class="inline-flex items-center align-middle justify-center font-medium border border-blue-50 bg-blue-50 text-blue-600 rounded-sm px-1 text-xs mr-1">
                                    Exit Node</div>
                            </span></div>
                    </div>
                </div>
            </header>
            <section class="mb-8">
                <header class="flex justify-between mb-4">
                    <div class="max-w-xl">
                        <h3 class="text-xl font-semibold tracking-tight mb-2">Subnets</h3>
                        <p class="text-gray-600">Subnets let you expose physical network routes onto Tailscale. <a
                                href="https://tailscale.com/kb/1019/subnets" target="_blank" rel="noopener noreferrer"
                                class="link"
                                aria-label="Read documentation about subnet routers">Learn&nbsp;more&nbsp;→</a></p>
                    </div>
                    <div><button class="button button-outline mt-2">Review</button></div>
                </header>
                <div class="p-4 md:p-6 border border-gray-200 rounded-md">
                    <ul class="leading-normal">
                        <li title="This IP is a subnet that has not been enabled." class="font-medium text-gray-400">
                            192.168.0.0/24</li>
                        <li title="This IP is a subnet that has not been enabled." class="font-medium text-gray-400">
                            192.168.1.0/24</li>
                        <li title="This IP is a subnet that has not been enabled." class="font-medium text-gray-400">
                            192.168.166.0/24</li>
                        <li title="This IP is a subnet that has not been enabled." class="font-medium text-gray-400">
                            192.168.168.0/24</li>
                        <li title="This IP is a subnet that has not been enabled." class="font-medium text-gray-400">
                            198.18.0.0/16</li>
                    </ul>
                </div>
            </section>
            <section class="mb-8">
                <header class="max-w-xl mb-4">
                    <h3 class="text-xl font-semibold tracking-tight mb-2">Machine Details</h3>
                    <p class="text-gray-600">Information about this machine’s network. Used to debug connection issues.
                    </p>
                </header>
                <div
                    class="p-4 md:p-6 border border-gray-200 rounded-md grid grid-cols-1 md:grid-cols-2 gap-y-2 sm:gap-x-2">
                    <div class="space-y-2">
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Creator</dt>
                            <dd class="min-w-0 truncate">gps949@nopkt.com</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Machine name</dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate">debian1520</div>
                                    <div class="cursor-pointer text-blue-500 pl-2">Copy</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Domain<span data-state="closed"><svg
                                        xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate">debian1520.cow-sole.ts.net</div>
                                    <div class="cursor-pointer text-blue-500 pl-2">Copy</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">OS hostname<span
                                    data-state="closed"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0 truncate">debian1520</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">OS</dt>
                            <dd class="min-w-0 truncate">Linux</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Tailscale version</dt>
                            <dd class="min-w-0 truncate">
                                <div class="flex items-center">1.34.2 </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Tailscale IPv4<span
                                    data-state="closed"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate"><span>100.90.174.53</span></div>
                                    <div class="cursor-pointer text-blue-500 pl-2">Copy</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Tailscale IPv6<span
                                    data-state="closed"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0">
                                <div class="flex relative min-w-0">
                                    <div class="truncate"><span
                                            class="inline-flex justify-start min-w-0 max-w-full"><span
                                                class="truncate w-fit flex-shrink">fd7a:115c:a1e0:ab12:4843</span><span
                                                class="flex-grow-0 flex-shrink-0">:cd96:625a:ae35</span></span></div>
                                    <div class="cursor-pointer text-blue-500 pl-2">Copy</div>
                                </div>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">ID<span data-state="closed"><svg
                                        xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0 truncate">n1kiTT3CNTRL</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Endpoints</dt>
                            <dd class="min-w-0 truncate">
                                <ul class="pl-3 -indent-3">
                                    <li class="select-all"><span>61.48.214.79</span><wbr>:<span>20332</span></li>
                                    <li class="select-all"><span>172.17.0.1</span><wbr>:<span>41641</span></li>
                                    <li class="select-all"><span>172.30.32.1</span><wbr>:<span>41641</span></li>
                                    <li class="select-all"><span>192.168.1.3</span><wbr>:<span>41641</span></li>
                                    <li class="select-all"><span>192.168.166.66</span><wbr>:<span>41641</span></li>
                                    <li class="select-all"><span>192.168.168.66</span><wbr>:<span>41641</span></li>
                                </ul>
                            </dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Relays<span data-state="closed"><svg
                                        xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0 truncate">
                                <ul>
                                    <li><strong class="font-medium">Relay #948</strong>: 160.70&nbsp;ms<svg
                                            xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-label="Preferred Relay"
                                            class="relative inline-block ml-1 -top-px">
                                            <polyline points="20 6 9 17 4 12"></polyline>
                                        </svg></li>
                                    <li><strong class="font-medium">Relay #951</strong>: 210.49&nbsp;ms</li>
                                </ul>
                            </dd>
                        </dl>
                    </div>
                    <div class="space-y-2">
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Created</dt>
                            <dd class="min-w-0 truncate">Aug 25, 2022 at 8:59 AM GMT+8</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Last seen</dt>
                            <dd class="min-w-0 truncate">8:27 PM GMT+8</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Key expiry</dt>
                            <dd class="min-w-0 truncate">No expiry</dd>
                        </dl>
                        <h3 class="pt-2 text-xs uppercase font-semibold text-gray-500 tracking-wide">Client connectivity
                        </h3>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Varies<span data-state="closed"><svg
                                        xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0 truncate">No</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">Hairpinning<span
                                    data-state="closed"><svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round"
                                        class="relative -top-px text-gray-500 hover:text-gray-800 ml-1 cursor-default inline-flex">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <line x1="12" y1="16" x2="12" y2="12"></line>
                                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                                    </svg></span></dt>
                            <dd class="min-w-0 truncate">No</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">IPv6</dt>
                            <dd class="min-w-0 truncate">No</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">UDP</dt>
                            <dd class="min-w-0 truncate">Yes</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">UPnP</dt>
                            <dd class="min-w-0 truncate">Yes</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">PCP</dt>
                            <dd class="min-w-0 truncate">Yes</dd>
                        </dl>
                        <dl class="flex text-sm">
                            <dt class="text-gray-500 w-1/3 md:w-1/4 mr-1 shrink-0">NAT-PMP</dt>
                            <dd class="min-w-0 truncate">No</dd>
                        </dl>
                    </div>
                </div>
            </section>
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
</template>

<style scoped>

</style>
