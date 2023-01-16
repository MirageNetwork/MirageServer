<script setup>
import { ref, computed, nextTick, onMounted } from "vue";

</script>

<template>
    <main class="container mx-auto pb-20 md:pb-24">
        <section class="mb-24">
            <header class="mb-8">
                <div class="flex justify-between items-center">
                    <div class="flex items-center">
                        <h1 class="text-3xl font-semibold tracking-tight leading-tight mb-2" tabindex="-1">
                            设备详情页 {{ $route.params.mip }}
                        </h1>
                    </div>
                </div>
            </header>

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
