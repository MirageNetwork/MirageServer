<script setup>
import { nextTick, onMounted, watch, ref } from 'vue';

const props = defineProps({
    show: Boolean,
    wxminicode: String,
    checkcode: String
})

const localWXMiniCode = ref("")
function checkWXMini() {
    return new Promise((resolve, reject) => {
        axios
            .post("/wxmini", {
                state: props.checkcode
            })
            .then(function (response) {
                // 处理成功情况
                resolve(response.data)
            })
            .catch(function (error) {
                // 处理错误情况
                reject("error");
            })
    });
}
onMounted(() => {
    localWXMiniCode.value = props.wxminicode
    const checkWXMiniIntID = setInterval(() => {
        checkWXMini().then(function (data) {
            switch (data["status"]) {
                case "Wait":
                    break;
                case "New":
                    localWXMiniCode.value = data["code"]
                    break;
                case "OK":
                    clearInterval(checkWXMiniIntID)
                    window.location.href = '/a/oauth_response?code=' + data["code"] + "&state=" + props.checkcode
                    break;
            }
        }).catch();
    }, 1500);
})
</script>

<template>
    <Transition enter-from-class="opacity-0 scale-y-75 " leave-to-class="opacity-0 scale-y-75"
        enter-active-class="transition ease-in-out duration-100 delay-150"
        leave-active-class="transition ease-in-out duration-100">
        <div v-if="show"
            class="bg-white rounded-md relative p-4 md:p-6 text-stone-700 max-w-sm min-w-[19rem] my-8 mx-auto w-[97%] shadow-md border-stone-200 border"
            style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">请使用微信扫描小程序码确认登录</div>
            </header>
            <h2>核验码: <strong class="text-red-400"> {{ checkcode.substring(3, 9) }} </strong></h2>
            <img :src="'data:image/jpeg;base64,' + localWXMiniCode" />
            <button @click="$emit('close')"
                class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
                type="button"><svg xmlns="http://www.w3.org/2000/svg" width="1.25em" height="1.25em" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg></button>
        </div>
    </Transition>
</template>