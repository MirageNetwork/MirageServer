<script setup>
import { nextTick, onMounted, ref } from 'vue';
import { useGetURLQuery } from '../utils'
import Register from './Register.vue'
import RegisterDone from './RegisterDone.vue'

const next_url = useGetURLQuery("next_url")

const showRegister = ref(false)
const closeRegister = ref(false)
const showRegSuccess = ref(false)
const regSuccessMsg = ref("")

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
</script>

<template>
    <div class="mb-10">
        <img class="h-8 w-24" src="/img/logo_withname@60.png" />
    </div>
    <form method="POST">
        <input type="hidden" name="provider" value="Ali">
        <input type="hidden" name="next_url" :value="next_url">
        <button type="submit"
            class="btn btn-outline rounded-md shadow border-stone-300 hover:border-stone-400 hover:bg-transparent text-black hover:text-black h-10 min-h-fit">
            <svg t="1674566173646" class="mr-3" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg"
                p-id="2782" width="20" height="20">
                <path
                    d="M959.2 383.9c-0.3-82.1-66.9-148.6-149.1-148.6H575.9l21.6 85.2 201 43.7c18.3 4.2 32.1 20.3 32.9 39.7 0.1 0.5 0.1 216.1 0 216.6-0.8 19.4-14.6 35.5-32.9 39.7l-201 43.7-21.6 85.3h234.2c82.1 0 148.8-66.5 149.1-148.6V383.9zM225.5 660.4c-18.3-4.2-32.1-20.3-32.9-39.7-0.1-0.6-0.1-216.1 0-216.6 0.8-19.4 14.6-35.5 32.9-39.7l201-43.7 21.6-85.2H213.8c-82.1 0-148.8 66.4-149.1 148.6V641c0.3 82.1 67 148.6 149.1 148.6H448l-21.6-85.3-200.9-43.9z m200.9-158.8h171v21.3h-171z"
                    fill="#ff7500" p-id="2783"></path>
            </svg>
            登录（阿里云IDaaS）
        </button>
    </form>
    <div class="mt-6 mb-2 text-stone-500 text-xs">还没有账号？</div>
    <Register :wantMeClose="closeRegister" :show="showRegister" @close="doCloseRegister" @reg-done="doRegSuccess">
    </Register>

    <Transition enter-from-class="opacity-0" enter-active-class="transition ease-in-out duration-75 delay-150">
        <button v-if="!showRegister && !showRegSuccess" @click="showRegister = true"
            class="btn rounded-md border-0 bg-stone-700 hover:bg-stone-800 h-10 min-h-fit mt-4">注册账号</button>
    </Transition>

    <RegisterDone :show="showRegSuccess" :welcomemsg="regSuccessMsg" @close="closeRegSuccess">
    </RegisterDone>

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
</template>

<style scoped>

</style>