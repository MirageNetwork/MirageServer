<script setup>
import { watch, ref, onMounted, computed } from 'vue';
import { useDisScroll } from '../utils.js';
import { useRouter, useRoute } from "vue-router";
import Toast from '../components/Toast.vue';

useDisScroll()
const router = useRouter();
const route = useRoute();

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
    if (toastShow.value) {
        setTimeout(function () { toastShow.value = false }, 5000)
    }
})

const userMenu = ref(null)
const props = defineProps({
    userName: String,
    userAccount: String,
    toleft: Number,
    totop: Number
})
const menuLeft = computed(() => {
    return String(String(props.toleft + 32 - userMenu.value?.clientWidth))
})
const menuTop = computed(() => {
    return String(props.totop + 40)
})

const emit = defineEmits(['close', 'logout'])
const closeMe = (event) => {
    emit('close')
}

function doLogout() {
    axios
        .get("/cockpit/api/logout")
        .then(function (response) {
            // 处理成功情况
            if (response.data["status"] == "success") {
                emit('logout')
                return
            }
            console.log("Get Logout Response ",response.data["status"])
        })
        .catch(function (error) {
            // 处理错误情况
            console.log(error);
        })
}
</script>

<template>
    <div ref="userMenu" class="shadow-lg border border-base-300 rounded-md z-20" v-click-away="closeMe"
        :style="'position: fixed; left: ' + menuLeft + 'px; top: ' + menuTop + 'px; min-width: max-content; --radix-popper-transform-origin: 100% 0px;'">
        <div class="dropdown bg-white rounded-md py-1 z-20"
            style="outline: none; --radix-dropdown-menu-content-transform-origin: var(--radix-popper-transform-origin); pointer-events: auto;">
            <div class="block px-4 py-2">
                <div><strong>{{ userName }} </strong></div>
                <div class="opacity-75">{{ userAccount }}</div>
            </div>
            <div class="my-1 border-b border-base-300"></div>
            <div class="relative block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100">
                <a class="stretched-link" @click="doLogout">登出</a>
            </div>
        </div>
    </div>
</template>

<style scoped></style>
