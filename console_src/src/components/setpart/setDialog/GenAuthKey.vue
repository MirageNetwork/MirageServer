<script setup>
import { watch, ref, onMounted, onBeforeUpdate, computed } from 'vue';
import { useDisScroll } from '/src/utils.js';

const emit = defineEmits(['added-authkey'])

useDisScroll()

const KeyGened = ref(false)
const inputBlocking = ref(false)

const isTagged = ref(false)
const isReusable = ref(false)
const isEphemeral = ref(false)


//输入框设置的密钥过期时长
const keyExpiryInputValue = ref(90);
const keyExpirySubDis = ref(false);
const keyExpiryAddDis = ref(true);

function updateKeyExpiryBtns() {
    if (Number(keyExpiryInputValue.value) > 1) {
        keyExpirySubDis.value = false;
    } else {
        keyExpirySubDis.value = true;
    }
    if (Number(keyExpiryInputValue.value) < 90) {
        keyExpiryAddDis.value = false;
    } else {
        keyExpiryAddDis.value = true;
    }
}
function keyExpiryCheck(isChange) {
    keyExpiryInputValue.value = keyExpiryInputValue.value
        .replace(/[^\d]+/g, "")
        .replace(/^0+(\d)/, "$1");
    //   if (keyExpiryInputValue.value == "") keyExpiryInputValue.value = 0;
    if (isChange) {
        if (keyExpiryInputValue.value == "") keyExpiryInputValue.value = 1;
        if (Number(keyExpiryInputValue.value) == 0) keyExpiryInputValue.value = 1;
        if (Number(keyExpiryInputValue.value) > 90) keyExpiryInputValue.value = 90;
    }
    updateKeyExpiryBtns();
}
function keyExpiryChange(isAdd) {
    if (isAdd == true) {
        keyExpiryInputValue.value = Number(keyExpiryInputValue.value) + 1;
    } else {
        keyExpiryInputValue.value = Number(keyExpiryInputValue.value) - 1;
    }
    updateKeyExpiryBtns();
}

function doKeyGen() {
    inputBlocking.value = true
    axios
        .post("/admin/api/keys", {
            keyData: {
                type: "authkey",
                expirySeconds: Number(keyExpiryInputValue.value * 24 * 3600),
                authkey: {
                    ephemeral: isEphemeral.value,
                    reusable: isReusable.value,
                    preauthorized: false
                }
            }
        })
        .then(function (response) {
            if (response.data["status"] == "success") {
                genedKey.value = response.data["data"]
                emit("added-authkey")
                KeyGened.value = true
            } else {
                console.log(response.data["status"])
            }
        })
        .catch(function (error) {
            console.log(error)
        })
        .then(function () {
            inputBlocking.value = false
        })
}

const copyBtnText = ref("复制");
const genedKey = ref({})

function copyGenedKey() {
    navigator.clipboard.writeText(genedKey.value["fullKey"]).then(function () {
        copyBtnText.value = "已复制!";
        setTimeout(() => {
            copyBtnText.value = "复制";
        }, 3000);
    });
}
</script>
<template>
    <div @click.self="$emit('close')" class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
        style="pointer-events: auto;">
        <div class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
            tabindex="-1" style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">生成授权密钥</div>
            </header>
            <form @submit.prevent="doKeyGen" v-if="!KeyGened">
                <div class="flex justify-between">
                    <div>
                        <h4 class="font-medium mb-1">可重用</h4>
                        <p class="text-sm text-gray-500">使用此密钥授权多个设备</p>
                    </div>
                    <div class="ml-6"><input :disabled="inputBlocking" v-model="isReusable" type="checkbox"
                            class="toggle"></div>
                </div>
                <div class="mt-4">
                    <h4 class="font-medium mb-1">过期</h4>
                    <p class="text-sm text-gray-500">该授权密钥有效期天数。这个并不会影响使用该密钥授权的设备自身的设备密钥有效期限</p>
                    <div class="flex mt-4">
                        <div class="relative">
                            <input :disabled="inputBlocking" v-model="keyExpiryInputValue"
                                @input="keyExpiryCheck(false)" @blur="keyExpiryCheck(true)"
                                class="input z-30 border focus:outline-blue-500/60 hover:border border-stone-200 hover:border-stone-400 rounded-r-none h-9 min-h-fit"
                                inputmode="numeric" pattern="[0-9]*" id="key-expiry-duration" tabindex="0" />
                            <div class="bg-white top-1 bottom-1 right-1 rounded-r-md absolute flex items-center">
                                <div class="flex items-center">
                                    <button @click="keyExpiryChange(false)"
                                        class="btn btn-ghost btn-sm px-2 hover:bg-stone-100 disabled:bg-transparent"
                                        :disabled="keyExpirySubDis" type="button" tabindex="-1">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round">
                                            <line x1="5" y1="12" x2="19" y2="12"></line>
                                        </svg></button><button @click="keyExpiryChange(true)"
                                        class="btn btn-ghost btn-sm px-2 hover:bg-stone-100 disabled:bg-transparent"
                                        :disabled="keyExpiryAddDis" type="button" tabindex="-1">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round">
                                            <line x1="12" y1="5" x2="12" y2="19"></line>
                                            <line x1="5" y1="12" x2="19" y2="12"></line>
                                        </svg>
                                    </button>
                                </div>
                            </div>
                        </div>
                        <div
                            class="flex items-center px-3 bg-gray-50 text-gray-500 border rounded-r border-l-0 border-gray-300">
                            天
                        </div>
                    </div>
                    <p class="text-sm text-gray-500 mt-1">必须是1-90天</p>
                </div>
                <div class="border-t font-medium text-sm text-gray-500 tracking-wider uppercase mt-6 mb-1 pt-6">设备设置
                </div>
                <p class="text-sm text-gray-500 mb-4">这些设置将应用于使用该密钥授权的所有设备</p>
                <div class="flex justify-between">
                    <div>
                        <h4 class="font-medium mb-1">自熄</h4>
                        <p class="text-sm text-gray-500">使用此密钥授权的设备会在离线时被自动从蜃境网络删除</p>
                    </div>
                    <div class="ml-6"><input :disabled="inputBlocking" v-model="isEphemeral" type="checkbox"
                            class="toggle"></div>
                </div>
                <!--
                <div class="flex justify-between mt-4">
                    <div>
                        <h4 class="font-medium mb-1">Pre-authorized</h4>
                        <p class="text-sm text-gray-500">Devices authenticated by this key will be automatically
                            authorized. <a href="https://tailscale.com/kb/1099/device-authorization/" target="_blank"
                                rel="noopener noreferrer" class="link"
                                aria-label="Read documentation about device authorization">Learn&nbsp;more&nbsp;→</a>
                        </p>
                    </div>
                    <div class="ml-6"><input type="checkbox" class="toggle"></div>
                </div>
                -->
                <div class="flex justify-between mt-4">
                    <div>
                        <h4 class="font-medium mb-1">标签</h4>
                        <p class="text-sm text-gray-500">使用该密钥授权的设备会自动添加标签。这同时会使该设备自身密钥过期被禁用
                        </p>
                    </div>
                    <div class="ml-6"><input :disabled="inputBlocking" v-model="isTagged" type="checkbox"
                            class="toggle"></div>
                </div>
                <div v-if="isTagged" class="rounded-md border border-stone-200 mt-4 mb-3 bg-stone-50 p-6">
                    <div class="flex justify-center">
                        <div class="w-full text-center max-w-xl text-gray-500">未分配标签</div>
                    </div>
                </div>
                <span v-if="isTagged">
                    <button
                        class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit"
                        disabled="" type="button">添加标签<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round" class="ml-1">
                            <polyline points="6 9 12 15 18 9"></polyline>
                        </svg></button>
                </span>
                <footer class="flex mt-10 justify-end space-x-4">
                    <button @click.self="$emit('close')"
                        class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit">取消</button>
                    <button type="submit"
                        class="btn border-0 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/60 text-white disabled:text-white/60 h-9 min-h-fit">生成密钥</button>
                </footer>
            </form>
            <form v-if="KeyGened">
                <p class="text-gray-700 mb-3">关闭前请确保您已复制下面新生成的密钥，它之后将不会再次完整展示</p>
                <div
                    class="flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0 mb-3 font-mono text-sm">
                    <input onclick="this.select()"
                        class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis" readonly
                        :value="genedKey.fullKey" />
                    <button @click="copyGenedKey"
                        class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap">
                        {{ copyBtnText }}
                    </button>
                </div>
                <div class="flex overflow-hidden rounded-md py-3 px-4 gap-2 text-sm bg-stone-50 text-gray-600 border border-stone-200"
                    role="alert">
                    <div class="pt-px"><svg xmlns="http://www.w3.org/2000/svg" width="1.125em" height="1.125em"
                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="16" x2="12" y2="12"></line>
                            <line x1="12" y1="8" x2="12.01" y2="8"></line>
                        </svg></div>
                    <div class="w-full">该密钥将在 {{ genedKey.expiry.split(' ')[0] }} 过期，之后您要想继续使用授权密钥需要重新生成</div>
                </div>
                <footer class="flex mt-10 justify-end space-x-4">
                    <button @click.self="$emit('close')"
                        class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit">完成</button>
                </footer>
            </form>
            <button @click="$emit('close')"
                class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
                type="button"><svg xmlns="http://www.w3.org/2000/svg" width="1.25em" height="1.25em" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg></button>
        </div>
    </div>
</template>
<style scoped>
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
    --togglehandleborder: 0 0 0 3px white inset, var(--handleoffsetcalculator) 0 0 3px white inset;
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