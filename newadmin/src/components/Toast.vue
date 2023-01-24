<script setup>
const props = defineProps({
    show: Boolean,
    msg: String
})
</script>

<template>
    <Transition enter-from-class="opacity-0 translate-y-6 scale-90"
        leave-to-class="opacity-0 translate-y-6 scale-90"
        enter-active-class="transition ease-in-out duration-300"
        leave-active-class="transition ease-in-out duration-300">
        <div v-if="show" >
            <div class="alert shadow-lg bg-neutral text-neutral-content rounded-md flex-row">
                <span>{{ msg }}</span>
                <svg @click="$emit('close')" cursor="pointer" xmlns="http://www.w3.org/2000/svg"
                    class="h-6 w-6 justify-self-end" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </div>
        </div>
    </Transition>
</template>

<style>

</style>