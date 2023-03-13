<script setup>
import { watch, ref, onMounted, nextTick, onBeforeUpdate, computed } from 'vue';
import { useDisScroll } from '/src/utils.js';

const emit = defineEmits(['update-done', 'update-fail', 'close'])

useDisScroll()

const inputBlocking = ref(false)

const props = defineProps({
    id: String,
    currentMachine: Object,
    tagOwners: Array,
    givenName: String,
})

const allTags = ref([])
const addedTags = ref([])
const containInvalidTags = computed(() => {
    return addedTags.value.some(tag => !allTags.value.includes(tag))
})
const activeBtn = ref(null)
const tagMenuLeft = ref(0)
const tagMenuTop = ref(0)
const tagMenuShow = ref(false)
function adjustTagMenuPosition() {
    if (activeBtn.value != null) {
        tagMenuLeft.value = activeBtn.value?.getBoundingClientRect().left
        tagMenuTop.value = activeBtn.value?.getBoundingClientRect().top - 8 - 40 * props.tagOwners.length
    }
}
function openTagMenu(event) {
    activeBtn.value = event.target
    while (activeBtn.value?.tagName != "BUTTON" && activeBtn.value?.tagName != "button") {
        activeBtn.value = activeBtn.value?.parentNode
    }
    adjustTagMenuPosition()
    tagMenuShow.value = true;
    nextTick(() => {
        window.addEventListener("scroll", adjustTagMenuPosition, true)
        window.addEventListener("resize", adjustTagMenuPosition, true)
    })
}
function closeTagMenu() {
    activeBtn.value = null
    tagMenuShow.value = false;
    nextTick(() => {
        window.removeEventListener("scroll", adjustTagMenuPosition, true)
        window.removeEventListener("resize", adjustTagMenuPosition, true)
    })
}
function addTag(tag) {
    closeTagMenu()
    if (addedTags.value.includes(tag)) {
        addedTags.value.splice(addedTags.value.indexOf(tag), 1)
    } else {
        addedTags.value.push(tag)
    }
}
onMounted(() => {
    addedTags.value = props.currentMachine.allowedTags.concat(props.currentMachine.invalidTags)
    allTags.value = []
    for (let i = 0; i < props.tagOwners.length; i++) {
        allTags.value.push(props.tagOwners[i]["tagName"])
    }
    console.log(allTags.value)
})


function updateTags() {
    inputBlocking.value = true
    axios
        .post("/admin/api/machines", {
            mid: props.id,
            state: "set-tags",
            tags: addedTags.value,
        })
        .then(function (response) {
            if (response.data["status"] == "success") {
                emit("update-done", props.id, response.data["data"]["allowedTags"], response.data["data"]["invalidTags"])
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
function isInvalidTag(tag) {
    for (var i in props.tagOwners) {
        if (props.tagOwners[i].tagName == tag) {
            return false
        }
    }
    return true
}
</script>

<template>
    <div @click.self="$emit('close')" class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
        style="pointer-events: auto;">
        <div class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
            style="pointer-events: auto;">
            <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
                <div class="font-semibold text-lg truncate">修改设备 {{ givenName }} 标签</div>
            </header>
            <form @submit.prevent="$emit('confirm')">
                <p class="text-gray-700 mb-6">标签可帮您实现不基于设备的创建者，而基于它的用途 (比如：<code
                        class="bg-gray-200 text-xs rounded px-1">server</code>)管理设备， 它们可以被用在ACL中控制访问</p>
                <div v-if="addedTags.length == 0"
                    class="rounded-md border border-stone-200 mt-4 mb-3 gap-2 bg-stone-50 p-6">
                    <div class="flex  justify-center">
                        <div class="w-full text-center max-w-xl text-gray-500">未分配标签</div>
                    </div>
                </div>
                <div v-if="addedTags.length > 0"
                    class="rounded-md border border-stone-200 mt-4 mb-3 flex flex-wrap gap-2 bg-stone-50 p-6">
                    <span v-for="tag, i in addedTags">
                        <div class="flex items-center align-middle justify-center font-medium border rounded-full px-2 py-1 leading-none text-xs"
                            :class="{
                                'border-gray-200 bg-gray-200 text-gray-600': isInvalidTag(tag),
                                'border-gray-300 bg-white': !isInvalidTag(tag),
                            }">
                            <svg v-if="isInvalidTag(tag)" xmlns="http://www.w3.org/2000/svg" width="10" height="10"
                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"
                                stroke-linecap="round" stroke-linejoin="round" class="mr-1 text-gray-500">
                                <circle cx="12" cy="12" r="10"></circle>
                                <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line>
                            </svg>
                            <span class="text-gray-500">{{ tag.substring(4) }}</span>
                            <span class="ml-1">
                                <button @click="addTag(tag)" type="button"><svg xmlns="http://www.w3.org/2000/svg"
                                        width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <line x1="18" y1="6" x2="6" y2="18"></line>
                                        <line x1="6" y1="6" x2="18" y2="18"></line>
                                    </svg></button>
                            </span>
                        </div>
                    </span>
                </div>
                <span>
                    <button :disabled="inputBlocking" @click="openTagMenu($event)"
                        class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit"
                        type="button">添加标签<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                            stroke-linejoin="round" class="ml-1">
                            <polyline points="6 9 12 15 18 9"></polyline>
                        </svg></button>
                </span>
                <div v-if="currentMachine.hasTags && addedTags.length == 0"
                    class="flex overflow-hidden rounded-md py-3 px-4 gap-2 text-sm mt-6 bg-orange-50 text-orange-800 border border-orange-100">
                    <div class="w-full">你不能保存修改，因为移除一个设备全部标签的唯一方式是重新进行授权。重授权该设备的用户将成为其管理者。</div>
                </div>
                <div v-if="containInvalidTags"
                    class="flex overflow-hidden rounded-md py-3 px-4 gap-2 text-sm mt-6 bg-orange-50 text-orange-800 border border-orange-100">
                    <div class="w-full">你不能保存修改，因为存在不可用的标签，您可以从设备上移除它们或者在ACL中定义它们。</div>
                </div>
                <footer class="flex mt-10 justify-end space-x-4">
                    <button :disabled="inputBlocking" @click="$emit('close')"
                        class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                        type="button">取消</button>
                    <button
                        :disabled="inputBlocking || containInvalidTags || currentMachine.hasTags && addedTags.length == 0"
                        @click="updateTags"
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
    <!--下方显示标签菜单-->
    <div v-if="tagMenuShow" v-click-away="closeTagMenu" class="shadow-xl border border-base-300 rounded-md z-20"
        :style="'position: fixed; left: 0px; top: 0px; transform: translate3d(' + tagMenuLeft + 'px, ' + tagMenuTop + 'px, 0px); min-width: max-content; z-index: 50; --radix-popper-transform-origin: 50% 155px;'">
        <div v-for="tag, i in tagOwners" class="bg-white rounded-md overflow-y-scroll max-h-80 max-w-xs z-50"
            style="outline: currentcolor; pointer-events: auto; --radix-dropdown-menu-content-transform-origin: var(--radix-popper-transform-origin);">
            <div @click="addTag(tag.tagName)"
                class="cursor-pointer hover:bg-stone-100 focus:outline-none focus:bg-bg-menu-item-hover border-b">
                <div class="h-full w-full flex justify-between items-center p-4 md:px-3 md:py-2">
                    <div class="w-6">
                        <span v-if="addedTags.includes(tag.tagName)" class="w-6">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <polyline points="20 6 9 17 4 12"></polyline>
                            </svg>
                        </span>
                    </div>
                    <div class="flex-1 text-gray-700 truncate">{{ tag.tagName.substring(4) }}</div>
                </div>
            </div>
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
