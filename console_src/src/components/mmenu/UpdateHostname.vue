<script setup>
import { watch, ref, onMounted, onBeforeUpdate, computed } from 'vue';
import { useDisScroll } from '/src/utils.js';

const emit = defineEmits(['update-done', 'update-fail', 'close'])

useDisScroll()

const inputBlocking = ref(false)

const setAutogen = ref(true)
const setName = ref("")
const setChanged = ref(false)
const valueChanged = ref(false)

const wantName = ref("")
const wantNameWithHash = ref(false)
const wantNameOccupied = ref(false)
const wrongSetName = ref(false)


const props = defineProps({
    id: String,
    hostName: String,
    givenName: String,
    autoGen: Boolean
})

watch(() => setAutogen.value, (newV) => {
    wantNameWithHash.value = false
    wantNameOccupied.value = false
    wrongSetName.value = false
    if (newV == props.autoGen) {
        setChanged.value = false
    } else {
        setChanged.value = true
    }
    if (newV) {
        if (props.autoGen) {
            setName.value = props.givenName
        } else {
            setName.value = props.hostName
        }
    }
})
watch(() => setName.value, (newV) => {
    wantNameWithHash.value = false
    wantNameOccupied.value = false
    wrongSetName.value = false
    setName.value = setName.value.toLowerCase().replace(/[^-0-9a-z]/ig, '-').replace(/--*/g, '-')
    if (newV == props.givenName) {
        valueChanged.value = false
    } else {
        valueChanged.value = true
    }
})

onMounted(() => {
    setChanged.value = false
    valueChanged.value = false
    setAutogen.value = props.autoGen
    setName.value = props.givenName
    wantNameWithHash.value = false
    wantNameOccupied.value = false
    wrongSetName.value = false
})


function updateName() {
    if (!(/^([0-9a-z]|-(?!-))+$/.test(setName.value))) {
        wrongSetName.value = true
        return
    }
    if (/^(-.*|.*-)$/.test(setName.value)) {
        wrongSetName.value = true
        return
    }
    inputBlocking.value = true
    var reqName = ""
    wantName.value = props.hostName
    if (!setAutogen.value) {
        reqName = setName.value
        wantName.value = setName.value
    }
    axios
        .post("/admin/api/machines", {
            mid: props.id,
            state: "rename-node",
            nodeName: reqName
        })
        .then(function (response) {
            if (response.data["status"] == "success") {
                if (response.data["data"]["name"] == wantName.value) {
                    emit("update-done", response.data["data"]["name"], response.data["data"]["automaticNameMode"], true)
                } else if (reqName === "") {
                    wantNameWithHash.value = true
                    emit("update-done", response.data["data"]["name"], response.data["data"]["automaticNameMode"], false)
                } else if (response.data["data"]["name"] == props.givenName) {
                    wantNameOccupied.value = true
                    emit("update-fail", "目标主机名已存在！")
                } else {
                    console.log("unexpected case happended during the hostname updating!")
                    emit("update-fail", "意外情形发生了！")
                }
            } else {
                emit("update-fail", response.data["status"].substring(6))
            }
        })
        .catch(function (error) {
            console.log(error);
            emit("update-fail", error)
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
                <div class="font-semibold text-lg truncate">修改设备名称 {{ givenName }}</div>
            </header>
            <form @submit.prevent="$emit('confirm')">
                <p class="text-gray-700 mb-6">这个名称显示在控制台、客户端，以及生成对应的「<strong>幻域</strong>」名称</p>
                <div class="flex items-center">
                    <input :disabled="inputBlocking" @change="autoGenToggle" v-model="setAutogen"
                        id="auto-generate-checkbox" type="checkbox" class="toggle toggle-sm">
                    <label for="auto-generate-checkbox" class="font-medium ml-3">由设备OS主机名生成</label>
                </div>
                <label for="machine-name" class="block font-medium mt-6 mb-2">设备名称</label>
                <div class="flex mb-2">
                    <div class="relative w-full z-30">
                        <input v-model="setName"
                            class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
                            type="text" :disabled="inputBlocking || setAutogen" id="machine-name">
                    </div>
                </div>

                <p v-if="wantNameWithHash" class="text-sm text-red-500 mb-2">名称 “{{ wantName }}” 已有设备占用 , 因此设备被自动命名为
                    “{{ setName }}”.</p>
                <p v-if="wantNameOccupied" class="text-sm text-red-500 mb-2">名称 “{{ wantName }}” 已有设备占用</p>
                <p v-if="wrongSetName" class="text-sm text-red-500 mb-2">名称不能为空，只能是字母数字和连接线组成，且不能以连接线开头结尾</p>

                <div v-if="!valueChanged" class="text-sm text-gray-600">这个设备可以通过「幻域」使用 <code
                        class="text-xs break-words bg-gray-200 px-1 rounded">{{ givenName }}</code> 进行访问</div>
                <div v-if="valueChanged" class="text-sm text-gray-600">这个设备可以通过「幻域」使用 <code
                        class="text-xs break-words bg-gray-200 px-1 rounded">{{ setName }}</code> 进行访问。 <code
                        class="text-xs break-words bg-gray-200 px-1 rounded">{{ givenName }}</code> 不再指向该设备。</div>
                <footer class="flex mt-10 justify-end space-x-4">
                    <button :disabled="inputBlocking" @click="$emit('close')"
                        class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                        type="button">取消</button>
                    <button :disabled="inputBlocking || !setChanged && !valueChanged" @click="updateName"
                        class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
                        type="submit">更新</button>
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
</style>
