<script setup>
import { watch, ref, onMounted, onBeforeUpdate, computed } from 'vue';
import { useDisScroll } from '/src/utils.js';

const emit = defineEmits(['added-tag',  'close'])

useDisScroll()

const inputBlocking = ref(false)

const setName = ref("")

const setNameOccupied = ref(false)
const wrongSetName = ref(false)


watch(() => setName.value, (newV) => {
    setNameOccupied.value = false
    wrongSetName.value = false
    setName.value = setName.value.toLowerCase().replace(/[^-0-9a-z]/ig, '-').replace(/--*/g, '-')
})

onMounted(() => {
    setNameOccupied.value = false
    wrongSetName.value = false
})


function createTagName() {
    if (!(/^([0-9a-z]|-(?!-))+$/.test(setName.value))) {
        wrongSetName.value = true
        return
    }
    if (/^(-.*|.*-)$/.test(setName.value)) {
        wrongSetName.value = true
        return
    }
    inputBlocking.value = true
    axios
        .post("/admin/api/tags", {
            state: "create",
            tagName: setName.value
        })
        .then(function (response) {
            if (response.data["status"] == "success") {
                    emit("added-tag")
                    emit("close")
            } else if (response.data["status"] == "error-occupied") { 
                setNameOccupied.value = true
            } else {
                console.log(response.data["status"])
            }
        })
        .catch(function (error) {
            console.log(error);
        });
    inputBlocking.value = false
}
</script>

<template>
    <div @click.self="$emit('close')" class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
        style="pointer-events: auto;">
        <div class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
            style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">创建标签</div>
            </header>
            <form @submit.prevent="$emit('confirm')">
                <p class="text-gray-700 mb-6">「<strong>标签</strong>」在被创建后可由管理员在控制台或对应标签管理员在登录时设置给节点</p>
                <label for="tag-name" class="block font-medium mt-6 mb-2">标签名称</label>
                <div class="flex mb-2">
                    <div class="relative w-full z-30">
                        <input v-model="setName"
                            class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
                            type="text" :disabled="inputBlocking" id="tag-name">
                    </div>
                </div>

                <p v-if="setNameOccupied" class="text-sm text-red-500 mb-2">标签名称 “{{ setName }}” 已存在</p>
                <p v-if="wrongSetName" class="text-sm text-red-500 mb-2">标签名称不能为空，只能是字母数字和连接线组成，且不能以连接线开头结尾</p>

                <footer class="flex mt-10 justify-end space-x-4">
                    <button :disabled="inputBlocking" @click="$emit('close')"
                        class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                        type="button">取消</button>
                    <button :disabled="inputBlocking" @click="createTagName"
                        class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
                        type="submit">创建</button>
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
</style>
