<script setup>
import { nextTick, onMounted, watch, ref } from 'vue';
import { useGetURLQuery } from '../utils'
import Register from './Register.vue'
import RegisterDone from './RegisterDone.vue'
import WxQR from './WxQR.vue'

import Toast from '../components/Toast.vue';

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
    if (toastShow.value) {
        setTimeout(function () { toastShow.value = false }, 5000)
    }
})

const next_url = useGetURLQuery("next_url")

const showRegister = ref(false)
const closeRegister = ref(false)
const showRegSuccess = ref(false)
const regSuccessMsg = ref("")

const WXMiniCode = ref("")
const WXMiniCheck = ref("")
const showWXMiniCode = ref(false)

onMounted(() => {
})
function doCloseRegister() {
    showRegister.value = false
    closeRegister.value = false
}
function doRegSuccess(data) {
    doCloseRegister()
    regSuccessMsg.value = data
    showRegSuccess.value = true
}
function closeRegSuccess() {
    regSuccessMsg.value = data
    showRegSuccess.value = false
}
function closeWXMini() {
    showWXMiniCode.value = false
}

function startWXScan() {
    axios
        .post("/login?provider=WXScan&next_url=" + next_url)
        .then(function (response) {
            if (response.data["status"] == "New") {
                WXMiniCode.value = response.data["code"]
                WXMiniCheck.value = response.data["state"]
                showWXMiniCode.value = true
            } else {
                toastMsg.value = "获取微信小程序码失败！"
                toastShow.value = true
            }
        })
        .catch(function (error) {
            reject(error)
        });
}

</script>

<template>
    <div class="mb-10">
        <img class="h-8 w-24" src="/img/logo_withname@60.png" />
    </div>
    <form v-if="!showWXMiniCode" method="POST">
        <input type="hidden" name="provider" value="Ali">
        <input type="hidden" name="next_url" :value="next_url">
        <button type="submit"
            class="btn btn-outline rounded-md shadow border-stone-300 hover:border-stone-400 hover:bg-transparent text-black hover:text-black h-10 min-h-fit w-64 justify-start">
            <svg t="1674566173646" class="ml-8 mr-3" viewBox="0 0 1024 1024" version="1.1"
                xmlns="http://www.w3.org/2000/svg" p-id="2782" width="20" height="20">
                <path
                    d="M959.2 383.9c-0.3-82.1-66.9-148.6-149.1-148.6H575.9l21.6 85.2 201 43.7c18.3 4.2 32.1 20.3 32.9 39.7 0.1 0.5 0.1 216.1 0 216.6-0.8 19.4-14.6 35.5-32.9 39.7l-201 43.7-21.6 85.3h234.2c82.1 0 148.8-66.5 149.1-148.6V383.9zM225.5 660.4c-18.3-4.2-32.1-20.3-32.9-39.7-0.1-0.6-0.1-216.1 0-216.6 0.8-19.4 14.6-35.5 32.9-39.7l201-43.7 21.6-85.2H213.8c-82.1 0-148.8 66.4-149.1 148.6V641c0.3 82.1 67 148.6 149.1 148.6H448l-21.6-85.3-200.9-43.9z m200.9-158.8h171v21.3h-171z"
                    fill="#ff7500" p-id="2783"></path>
            </svg>
            登录（阿里云IDaaS）
        </button>
    </form>
    <form v-if="!showWXMiniCode" @submit.prevent="startWXScan" class="mt-3">
        <input type="hidden" name="provider" value="WXScan">
        <input type="hidden" name="next_url" :value="next_url">
        <button type="submit"
            class="btn btn-outline rounded-md shadow border-stone-300 hover:border-stone-400 hover:bg-transparent text-black hover:text-black h-10 min-h-fit w-64 justify-start">
            <svg xmlns="http://www.w3.org/2000/svg" class="ml-8 mr-3" viewBox="0 0 24 24" width="20" height="20">
                <path fill="none" d="M0 0h24v24H0z" />
                <path
                    d="M15.84 12.691l-.067.02a1.522 1.522 0 0 1-.414.062c-.61 0-.954-.412-.77-.921.136-.372.491-.686.925-.831.672-.245 1.142-.804 1.142-1.455 0-.877-.853-1.587-1.905-1.587s-1.904.71-1.904 1.587v4.868c0 1.17-.679 2.197-1.694 2.778a3.829 3.829 0 0 1-1.904.502c-1.984 0-3.598-1.471-3.598-3.28 0-.576.164-1.117.451-1.587.444-.73 1.184-1.287 2.07-1.541a1.55 1.55 0 0 1 .46-.073c.612 0 .958.414.773.924-.126.347-.466.645-.861.803a2.162 2.162 0 0 0-.139.052c-.628.26-1.061.798-1.061 1.422 0 .877.853 1.587 1.905 1.587s1.904-.71 1.904-1.587V9.566c0-1.17.679-2.197 1.694-2.778a3.829 3.829 0 0 1 1.904-.502c1.984 0 3.598 1.471 3.598 3.28 0 .576-.164 1.117-.451 1.587-.442.726-1.178 1.282-2.058 1.538zM2 12c0 5.523 4.477 10 10 10s10-4.477 10-10S17.523 2 12 2 2 6.477 2 12z"
                    fill="rgba(56,186,109,1)" />
            </svg>
            登录（微信小程序扫码）
        </button>
    </form>
    <div v-if="!showRegister && !showRegSuccess && !showWXMiniCode" class="mt-6 mb-2 text-stone-500 text-xs">还没有账号？</div>
    <Register :wantMeClose="closeRegister" :show="showRegister" @close="doCloseRegister" @reg-done="doRegSuccess">
    </Register>

    <Transition enter-from-class="opacity-0" enter-active-class="transition ease-in-out duration-75 delay-150">
        <button v-if="!showRegister && !showRegSuccess && !showWXMiniCode" @click="showRegister = true"
            class="btn rounded-md border-0 bg-stone-700 hover:bg-stone-800 h-10 min-h-fit mt-4">注册账号</button>
    </Transition>

    <RegisterDone :show="showRegSuccess" :welcomemsg="regSuccessMsg" @close="closeRegSuccess">
    </RegisterDone>

    <WxQR v-if="showWXMiniCode" :show="showWXMiniCode" :wxminicode="WXMiniCode" :checkcode="WXMiniCheck" @close="closeWXMini">
    </WxQR>

    <footer class="mt-10 text-sm text-stone-600">
        <p><strong>不会用？</strong> 了解更多请发邮件至<a class="underline" href="mailto:gps949@nopkt.com?subject=[关于蜃境]"> gps949
                (AT) nopkt.com </a>.
        </p>
    </footer>
    <footer class="mt-16 max-w-md text-sm text-stone-600">
        <p>点击以上按钮进行操作，表明您已经阅读、理解并同意蜃境网络的 <br />
            <a class="underline" href="#" target="_blank">服务条款</a> 以及 <a class="underline" href="#"
                target="_blank">隐私策略</a>.
        </p>
    </footer>
    <!-- 提示框显示 -->
    <Teleport to=".toast-container">
        <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
    </Teleport>
</template>

<style scoped></style>