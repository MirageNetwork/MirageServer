<script setup>
import { watch, ref, onMounted, computed } from 'vue';
import { useDisScroll } from '../utils.js';

useDisScroll()

const props = defineProps({
    machineName: String
})
</script>

<template>
    <div @click.self="$emit('close')" class="fixed overflow-y-auto inset-0 py-8 z-10 bg-gray-900 bg-opacity-[0.07]"
        style="pointer-events: auto;">
        <div class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
            style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">删除 {{ machineName }}</div>
            </header>
            <form @submit.prevent="$emit('confirm')">
                <p class="text-gray-700 mb-4">这个设备将从您的蜃境网络中永久删除！如需重新添加该设备，您将需要在该设备上重新进行授权</p>
                <footer class="flex mt-10 justify-end space-x-4">
                    <button @click="$emit('close')"
                        class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                        type="button">取消</button>
                    <button class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit"
                        type="submit">删除设备</button>
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

</style>
