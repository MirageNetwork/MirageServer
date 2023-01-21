<script setup>
import { watch, ref, onMounted, computed } from 'vue';
import { useDisScroll } from '../utils.js';

useDisScroll()

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

const emit = defineEmits(['close'])
const closeMe = (event) => {
    emit('close')
}

</script>

<template>
    <div ref="userMenu" class="shadow-lg border-b border-gray-200" v-click-away="closeMe"
        :style="'position: fixed; left: ' + menuLeft + 'px; top: ' + menuTop + 'px; min-width: max-content; z-index: 50; --radix-popper-transform-origin: 100% 0px;'">
        <div class="dropdown bg-white rounded-md py-1 z-50"
            style="outline: none; --radix-dropdown-menu-content-transform-origin: var(--radix-popper-transform-origin); pointer-events: auto;">
            <div class="block px-4 py-2">
                <div><strong>{{ userName }} </strong></div>
                <div class="opacity-75">{{ userAccount }}</div>
            </div>
            <div class="my-1 border-b border-gray-200"></div>
            <div
                class="relative block px-4 py-2 cursor-pointer hover:enabled:bg-gray-100 focus:outline-none focus:bg-gray-100">
                <a class="stretched-link" href="/admin/logout">登出</a>
            </div>
        </div>
    </div>
</template>

<style scoped>

</style>
