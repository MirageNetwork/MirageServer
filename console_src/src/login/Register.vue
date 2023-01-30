<script setup>
import { onMounted, onBeforeUpdate, watch, ref, nextTick } from 'vue';
import Identify from '../Identify.vue';
import VerifyCode from '../VerifyCode.vue';
import Toast from '../components/Toast.vue';

const emit = defineEmits(['close', 'reg-done'])

const props = defineProps({
    show: Boolean,
    wantMeClose: Boolean
})

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
    if (toastShow.value) {
        setTimeout(function () { toastShow.value = false }, 5000)
    }
})

const nameInput = ref(null)
const phoneInput = ref(null)
const captchaInput = ref("")
const codeInput = ref("")
const codeBtnText = ref("获取验证码")

const nameTipShow = ref(false)
const phoneTipShow = ref(false)
const notifyName = ref(false)
const notifyPhone = ref(false)

const sendCodeBtnDis = ref(true)
const regBtnDis = ref(true)


const preventAction = ref(false)

const focusOnCode = ref(false)
const idCodes = "0123456789abcdefghijklmnopqrstuvwxyz"
const idCode = ref("jida")
function makeCode(len) {
    if (sendCodeBtnDis.value && regBtnDis.value) {
        idCode.value = ""
        for (let i = 0; i < len; i++) {
            idCode.value += idCodes[Math.floor(Math.random() * idCodes.length)]
        }
    }
}

onMounted(() => {
    makeCode(4)
})

function closeMe() {
    clearInterval(codeCounterID.value)
    nameTipShow.value = false
    phoneTipShow.value = false
    sendCodeBtnDis.value = true
    regBtnDis.value = true
    makeCode(4)
    captchaInput.value = ""
    codeInput.value = ""
    codeBtnText.value = "获取验证码"

    emit('close')
}

watch(() => props.wantMeClose, (newV) => {
    if (newV) {
        closeMe()
    }
})

onBeforeUpdate(() => {
    /*
    clearInterval(codeCounterID.value)
    nameTipShow.value = false
    phoneTipShow.value = false
    sendCodeBtnDis.value = true
    regBtnDis.value = true
    makeCode(4)
    captchaInput.value = ""
    codeInput.value = ""
    codeBtnText.value = "获取验证码"
    */
})

const specialReg = RegExp(
    /[(\ )(\`)(\~)(\!)(\@)(\#)(\$)(\%)(\^)(\&)(\*)(\()(\))(\-)(\_)(\+)(\=)(\[)(\])(\{)(\})(\|)(\\)(\;)(\:)(\')(\")(\,)(\.)(\/)(\<)(\>)(\?)(\)(\·)(\！)(\￥)(\…)(\（)(\）)(\——)(\【)(\】)(\「)(\」)(\、)(\；)(\：)(\“)(\”)(\’)(\‘)(\？)(\《)(\》)(\，)(\。)]+/
)
const phoneReg = /^[1][3,4,5,6,7,8,9][0-9]{9}$/
const codeReg = /^[0-9]{0-6}$/

function nameCheck() {
    if (specialReg.test(nameInput.value?.value)) {
        nameTipShow.value = true
    } else {
        nameTipShow.value = false
        notifyName.value = false
    }
    if (nameInput.value?.value == "") {
        nameTipShow.value = true
    }
}
function phoneCheck() {
    if (phoneReg.test(phoneInput.value?.value)) {
        phoneTipShow.value = false
        notifyPhone.value = false
    } else {
        phoneTipShow.value = true
    }
}
function captchaCheck(event) {
    if (captchaInput.value.length == idCode.value.length) {
        if (captchaInput.value != idCode.value) {
            captchaInput.value = ""
            makeCode(4)
        } else {
            event.target.blur()
            sendCodeBtnDis.value = false
        }
    }
}
function getNewVerifyCode(inputedCode) {
    if (inputedCode.length == 0) {
        codeInput.value = ""
        //        console.log("不再合规！")
    } else {
        codeInput.value = inputedCode
        //        console.log("输入完一个验证码：" + inputedCode)
    }
}
/*
watch(() => codeInput.value, (newV, oldV) => {
    console.log(newV + codeReg.test(newV))
    if (!codeReg.test(newV)) {
        codeInput.value = oldV
    }
})
*/

const codeCounterID = ref()
function startCodeCounter() {
    let counter = 5
    sendCodeBtnDis.value = true
    codeBtnText.value = "重新获取(" + counter + "s)"
    codeCounterID.value = setInterval(() => {
        if (counter > 0) {
            counter--
            codeBtnText.value = "重新获取(" + counter + "s)"
            if (counter == 0) {
                clearInterval(codeCounterID.value)
                codeBtnText.value = "重新获取"
                sendCodeBtnDis.value = false
            }
        }
    }, 1000)
}

function reqSMSCode() {
    return new Promise((resolve, reject) => {
        axios
            .post("/api/register", {
                name: nameInput.value?.value,
                mobile: phoneInput.value?.value,
            })
            .then(function (response) {
                if (response.data["status"] == "success") {
                    resolve("success")
                } else {
                    reject(response.data["status"].substring(6))
                }
            })
            .catch(function (error) {
                reject(error)
            });
    });
}
function sendCode() {
    nameCheck()
    phoneCheck()
    if (nameTipShow.value || phoneTipShow.value) {
        if (nameTipShow.value) {
            notifyName.value = true
            setTimeout(() => {
                notifyName.value = false
            }, 3000)
        }
        if (phoneTipShow.value) {
            notifyPhone.value = true
            setTimeout(() => {
                notifyPhone.value = false
            }, 3000)
        }
    } else {
        reqSMSCode().then(function (msg) {
            regBtnDis.value = false;
            startCodeCounter();
        }).catch(function (err) {
            toastMsg.value = err;
            toastShow.value = true;
        })
    }
}
function reqReg() {
    return new Promise((resolve, reject) => {
        axios
            .post("/api/register", {
                name: nameInput.value?.value,
                mobile: phoneInput.value?.value,
                verifyCode: codeInput.value
            })
            .then(function (response) {
                if (response.data["status"] == "success") {
                    resolve(response)
                } else {
                    reject(response.data["status"].substring(6))
                }
            })
            .catch(function (error) {
                reject(error)
            });
    })
}
function doRegister() {
    preventAction.value = true
    reqReg().then((res) => {
        preventAction.value = false
        emit('reg-done', res.data["data"])
    }).catch((err) => {ccbejnchjvrvcdtlrhgbfbhncivnnlvjuleiighnr
        
        preventAction.value = false
        toastMsg.value = err;
        toastShow.value = true;
    })
}
</script>

<template>
    <Transition enter-from-class="opacity-0 scale-y-75 " leave-to-class="opacity-0 scale-y-75"
        enter-active-class="transition ease-in-out duration-100 delay-150"
        leave-active-class="transition ease-in-out duration-100">
        <div v-if="show"
            class="bg-white rounded-md relative p-4 md:p-6 text-stone-700 max-w-sm min-w-[19rem] my-8 mx-auto w-[97%] shadow-md border-stone-200 border"
            style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">注册新用户</div>
            </header>
            <p class="flex justify-start text-stone-500 mt-2 mb-8  text-xs text-left">*测试阶段仅支持手机号注册</p>
            <form @submit.prevent="$emit('confirm')">
                <input type="hidden" id="fp" name="fp" />
                <div
                    class="flex flex-row mb-5 border rounded-md border-stone-300 hover:border-stone-400 max-w-xs tooltip-bottom tooltip-error">
                    <div
                        class="flex flex-row rounded-l-md rounded-r-none bg-stone-200 text-base align-middle content-center h-9 w-16 -mr-16">
                        <svg t="1674637147479" class="icon h-9 w-9 -mr-9" viewBox="0 0 1024 1024" version="1.1"
                            xmlns="http://www.w3.org/2000/svg" p-id="3031">
                            <path
                                d="M511.991302 128.08519c-212.038254 0-383.906623 171.883719-383.906623 383.91481s171.868369 383.91481 383.906623 383.91481c212.031091 0 383.922996-171.883719 383.922996-383.91481S724.022393 128.08519 511.991302 128.08519zM511.991302 847.925842c-79.240739 0-151.927201-27.603685-209.382774-73.499991l158.574598-84.949764-37.632085-29.79049c-46.208415-36.601616-63.438844-113.00575-63.438844-152.559605L360.112196 406.961712c0-45.36521 75.061556-108.476597 148.966777-108.476597 74.154907 0 146.654105 62.111616 146.654105 108.476597l0 100.165304c0 39.646975-12.450567 116.412336-59.143005 152.747893l-38.601156 30.008454 162.253384 85.386716C662.981155 820.635289 590.731644 847.925842 511.991302 847.925842z"
                                fill="#78716c" p-id="3032"></path>
                        </svg>
                    </div>
                    <input ref="nameInput" :disabled="!regBtnDis" @input="nameCheck" type="text" name="name"
                        placeholder="请输入您的姓名"
                        class="border-0 rounded-md bg-transparent w-full h-9 pr-3 pl-[4.5rem] leading-5 placeholder:text-sm text-sm" />
                    <svg v-if="notifyName" t="1674704280252" class="icon h-9 -mr-16 animation-here"
                        viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="3838">
                        <path
                            d="M689.664 917.504L290.816 552.96c-12.288-10.752-19.456-26.624-19.456-43.008s7.168-32.256 19.456-43.008L689.664 102.4c18.944-17.92 46.08-23.552 70.656-14.336 24.064 9.216 40.96 31.232 43.52 57.344v729.088c-2.048 26.112-18.944 48.128-43.52 57.344-24.576 9.216-51.712 3.584-70.656-14.336z"
                            fill="#ef4444" p-id="3839"></path>
                    </svg>
                </div>
                <div v-if="nameTipShow" class="flex text-xs text-red-500 -mt-4 justify-end">
                    <p>请输入合法用户名</p>
                </div>
                <div
                    class="flex flex-row justify-self-end mb-5 border rounded-md border-stone-300 hover:border-stone-400 max-w-xs">
                    <div
                        class="flex flex-row rounded-l-md rounded-r-none bg-stone-200 text-base align-middle content-center h-9 w-16 -mr-16">
                        <svg t="1674636847709" class="icon h-9 w-9" viewBox="0 0 1024 1024" version="1.1"
                            xmlns="http://www.w3.org/2000/svg" p-id="3935">
                            <path
                                d="M693.352883 126.340453 329.308633 126.340453c-44.685735 0-80.910774 36.22197-80.910774 80.901565l0 606.755084c0 44.677549 36.224016 80.908728 80.910774 80.908728l364.04425 0c44.676525 0 80.901565-36.231179 80.901565-80.908728L774.254447 207.240995C774.254447 162.562423 738.030431 126.340453 693.352883 126.340453zM572.002071 834.222749 450.650235 834.222749l0-40.450271 121.350812 0L572.001047 834.222749zM713.57853 733.097584c0 11.156084-9.046027 20.224624-20.225647 20.224624L329.308633 753.322208c-11.178597 0-20.225647-9.067516-20.225647-20.224624L309.082986 207.240995c0-11.180644 9.04705-20.225647 20.225647-20.225647l364.04425 0c11.17962 0 20.225647 9.045003 20.225647 20.225647L713.57853 733.097584z"
                                fill="#78716c" p-id="3936"></path>
                        </svg>
                        <svg t="1674641827678" class="icon h-9 w-4" viewBox="0 0 1024 1024" version="1.1"
                            xmlns="http://www.w3.org/2000/svg" p-id="4249">
                            <path
                                d="M427.469 451.738l55.091 61.747-73.165 58.726H160.922c-13.568 0-22.58 0-27.239-1.536-13.517 0-30.771-7.526-38.963-15.053L56.525 514.97l68.25-63.284 302.694 0.052z m91.443 512L453.427 1024l-54.937-60.262 33.587-373.504 64.256-57.242 56.217 57.242-33.638 373.504z m45.875-525.568l-65.485 60.262-53.606-61.747 32.205-371.917L544.563 0l53.863 64.768-33.639 373.402z m0.307 136.243l-56.115-60.877 72.243-59.904 248.423-3.994c13.568-0.204 22.579-0.358 27.29 1.076 13.516-0.205 30.873 7.014 39.219 14.438l38.81 40.038-67.226 64.359-302.644 4.864z"
                                p-id="4250" fill="#78716c"></path>
                        </svg>
                        <svg t="1674641867132" class="icon h-9 w-5" viewBox="0 0 1024 1024" version="1.1"
                            xmlns="http://www.w3.org/2000/svg" p-id="944">
                            <path
                                d="M303.95 890.37l-75.717 64.503-29.169-35.334c-8.477-9.21-14.18-21.502-15.374-35.296-0.54-6.165 0.346-13.795 1.194-21.501l19.652-231.812c2.735-21.5 8.9-38.416 21.655-50.631l52.52-49.13 52.288 53.715-27.05 305.485z m40.227-452.872l-64.118 56.796-30.44-32.213c-13.525-13.794-22.773-32.251-24.006-46.046-0.655-7.668 0.077-16.877 0.655-27.666l18.573-208.77c2.042-29.168 8.63-41.422 29.785-62.923l37.993-39.92 57.953 66.007-26.395 294.735z m289.263 463.66l61.614 72.172-31.366 27.628c-21.27 19.96-30.248 23.042-70.129 23.042H337.164c-38.378 0-54.138-4.624-69.55-23.042l-23.775-26.125 82.613-73.674H633.44z m40.46-449.827l56.18 62.962-68.434 59.84H356.122l-57.452-59.879 68.203-62.962h307.026zM393.421 122.84l-60.92-64.503 30.941-32.252C381.67 6.126 394.963 0 421.05 0h277.858c24.584 0 38.918 6.127 54.485 26.086l28.822 30.75-80.224 66.005H393.422z m285.294 471.252l68.318-61.382 33.793 35.296c10.134 10.75 15.952 24.545 17.301 39.92 0.81 9.17 0.077 18.418-0.54 29.169l-20.036 227.187c-0.463 12.292-2.736 21.501-3.892 26.086-2.273 9.21-7.707 16.916-23.39 30.71l-38.378 35.335-59.494-66.045 26.318-296.276z m40.228-452.872l78.683-66.006 25.547 29.17c10.288 12.291 15.953 24.583 17.032 36.836 0.655 7.668-0.193 15.336-0.925 24.584L820.515 389.91c-4.315 38.378-12.06 55.294-33.446 73.713l-36.952 33.754-57.452-59.88 26.279-296.276z"
                                p-id="945" fill="#78716c"></path>
                        </svg>
                        <svg t="1674641902866" class="icon h-9 w-5" viewBox="0 0 1024 1024" version="1.1"
                            xmlns="http://www.w3.org/2000/svg" p-id="1130">
                            <path
                                d="M326.206 885.783l-75.562 66.006-27.28-30.71c-11.946-13.756-17.495-24.545-18.574-36.837-0.54-6.165 0.04-16.877 0.77-26.125l23.12-262.522 70.977-66.006 52.558 56.797-26.01 299.397z m40.498-449.827l-65.506 58.338-57.336-58.338 22.542-268.648c3.467-30.71 8.515-43.002 29.94-61.42l37.068-32.252 59.378 64.503-26.086 297.817z m283.252 465.203l63.116 72.17-25.624 23.005c-25.624 23.042-39.034 27.666-82.036 27.666H372.06c-50.67 0-58.608-3.083-80.687-27.666L269.41 973.33l79.647-72.17h300.9z m9.787-449.827c30.71 0 40.15 3.044 60.61 26.125l29.17 35.333-69.86 61.344h-302.44l-57.568-61.383 68.318-61.458h271.77z m-245.22-328.49L352.06 58.338l35.412-33.754C407.393 6.127 422.19 0 459.027 0H752.26l60.496 59.88-71.246 62.96H414.522z m286.45 466.666l65.659-56.797 35.45 36.837c11.83 12.254 15.567 19.96 16.76 33.755 0.81 9.21 0.888 27.666-0.731 44.505L800.54 850.45c-4.47 36.875-10.79 52.211-32.098 72.171l-35.411 33.793-59.65-67.586 27.59-299.32z"
                                p-id="1131" fill="#78716c"></path>
                        </svg>
                    </div>
                    <input ref="phoneInput" :disabled="!regBtnDis" @input="phoneCheck" type="text" name="mobile"
                        placeholder="请输入您的手机号"
                        class="border-0 rounded-md bg-transparent w-full h-9 pr-3 pl-[4.5rem] leading-5 placeholder:text-sm text-sm" />
                    <svg v-if="notifyPhone" t="1674704280252" class="icon h-9 -mr-16 animation-here"
                        viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="3838">
                        <path
                            d="M689.664 917.504L290.816 552.96c-12.288-10.752-19.456-26.624-19.456-43.008s7.168-32.256 19.456-43.008L689.664 102.4c18.944-17.92 46.08-23.552 70.656-14.336 24.064 9.216 40.96 31.232 43.52 57.344v729.088c-2.048 26.112-18.944 48.128-43.52 57.344-24.576 9.216-51.712 3.584-70.656-14.336z"
                            fill="#ef4444" p-id="3839"></path>
                    </svg>
                </div>
                <div v-if="phoneTipShow" class="flex text-xs text-red-500 -mt-4 justify-end">
                    <p>请输入合法手机号</p>
                </div>
                <div class="flex flex-row items-center justify-self-end mb-5 border rounded-md max-w-xs"
                    :class="{ 'border-stone-300 hover:border-stone-400': sendCodeBtnDis && codeBtnText == '获取验证码', 'border-green-400': !sendCodeBtnDis || codeBtnText != '获取验证码' }">
                    <input v-model="captchaInput" :disabled="!sendCodeBtnDis || !regBtnDis && captchaInput == idCode"
                        @input="captchaCheck" type="text" name="captcha" placeholder="请输入图形验证码" autocomplete="off"
                        class="border-0 rounded-md w-full h-9 px-3 leading-5 pr-24 placeholder:text-sm text-sm" />
                    <svg v-if="!sendCodeBtnDis || codeBtnText != '获取验证码'" t="1674701202328" class="icon h-9 mr-20"
                        viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="7795">
                        <path
                            d="M802.922882 383.309012 428.076612 758.155283 220.943065 551.154765c-22.317285-22.317285-22.317285-55.993269 0-78.310553 22.450315-22.450315 55.993269-22.450315 78.443583 0l128.689964 128.689964L724.613352 304.999482c22.450315-22.450315 55.993269-22.450315 78.30953 0C825.373197 327.316767 825.373197 360.858698 802.922882 383.309012zM512 64.322981c-246.155283 0-447.677019 201.521736-447.677019 447.677019s201.521736 447.677019 447.677019 447.677019 447.677019-201.521736 447.677019-447.677019S758.155283 64.322981 512 64.322981z"
                            fill="#4ade80" p-id="7796"></path>
                    </svg>
                    <Identify @click="makeCode(4)" :identifyCode="idCode"></Identify>

                </div>
                <div class="flex flex-row justify-self-end mb-5 rounded-md max-w-xs">
                    <!--  <input v-model="codeInput" :disabled="regBtnDis" type="text" name="code" placeholder="请输入短信验证码"
                        autocomplete="off"
                        class="border-0 rounded-md w-full h-9 px-3 leading-5 pr-20 placeholder:text-sm text-sm" />
                    -->
                    <VerifyCode :switchoff="regBtnDis" :setFocus="focusOnCode" @focusSet="focusOnCode = false"
                        @complete="getNewVerifyCode">
                    </VerifyCode>
                    <button ref="sendCodeBtn" :disabled="sendCodeBtnDis" @click="sendCode"
                        class="text-xs h-9 w-20 z-10 text-blue-500 hover:text-blue-900 disabled:text-stone-500/60">{{
                            codeBtnText
                        }}</button>
                </div>


                <footer class="flex mt-10 justify-end space-x-4">
                    <button @click="closeMe"
                        class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
                        type="button">取消</button>
                    <button ref="regBtn" :disabled="regBtnDis || codeInput.length < 6" @click="doRegister"
                        class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
                        type="button">注册</button>
                </footer>
            </form>
            <button @click="closeMe"
                class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
                type="button"><svg xmlns="http://www.w3.org/2000/svg" width="1.25em" height="1.25em" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg></button>
        </div>
    </Transition>
    <div v-if="preventAction" class="flex fixed overflow-y-auto inset-0 py-8 z-20 bg-gray-900 bg-opacity-[0.07]"
        style="pointer-events: auto;">
        <svg t="1674733336391" class="icon w-full animate-slowspin self-center" viewBox="0 0 1024 1024" version="1.1"
            xmlns="http://www.w3.org/2000/svg" p-id="9256" width="200" height="200">
            <path d="M448 224a96 64 90 1 0 128 0 96 64 90 1 0-128 0Z" fill="#57534e" opacity=".9" p-id="9257"></path>
            <path d="M448 800a96 64 90 1 0 128 0 96 64 90 1 0-128 0Z" fill="#57534e" opacity=".5" p-id="9258"></path>
            <path d="M704 512a96 64 0 1 0 192 0 96 64 0 1 0-192 0Z" fill="#57534e" opacity=".7" p-id="9259"></path>
            <path d="M128 512a96 64 0 1 0 192 0 96 64 0 1 0-192 0Z" fill="#57534e" opacity=".3" p-id="9260"></path>
            <path
                d="M647.766905 374.606262a64 96 44.999 1 0 135.762133-135.766872 64 96 44.999 1 0-135.762133 135.766872Z"
                fill="#57534e" opacity=".8" p-id="9261"></path>
            <path
                d="M240.470962 785.16061a64 96 44.999 1 0 135.762133-135.766872 64 96 44.999 1 0-135.762133 135.766872Z"
                fill="#57534e" opacity=".4" p-id="9262"></path>
            <path d="M672.02313 760.903595a96 64 44.999 1 0 90.508088-90.511247 96 64 44.999 1 0-90.508088 90.511247Z"
                fill="#57534e" opacity=".6" p-id="9263"></path>
            <path d="M261.468782 353.607652a96 64 44.999 1 0 90.508088-90.511247 96 64 44.999 1 0-90.508088 90.511247Z"
                fill="#57534e" opacity=".2" p-id="9264"></path>
        </svg>
    </div>

    <!-- 提示框显示 -->
    <Teleport to=".toast-container">
        <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
    </Teleport>
</template>

<style scoped>
.animation-here {
    animation:
        bounce 1s infinite;
}

@keyframes bounce {

    0%,
    100% {
        transform: translateX(-25%);
        animation-timing-function: cubic-bezier(0.8, 0, 1, 1);
    }

    50% {
        transform: translateX(0);
        animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
    }
}

.animate-slowspin {
    animation: spin 2s linear infinite;
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }

    to {
        transform: rotate(360deg);
    }
}
</style>