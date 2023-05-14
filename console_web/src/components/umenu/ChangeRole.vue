<script setup>
import { ref, watch, computed } from "vue";
import { useDisScroll } from "/src/utils.js";

const emit = defineEmits(["change-role"]);

useDisScroll();

const props = defineProps({
  selectUser: Object,
  wantedRole: String,
  canAssignOwner: Boolean,
});

const confirmTransferOwner = ref(false);

function goSubmit() {
  if (props.wantedRole === "owner" && !confirmTransferOwner.value) {
    confirmTransferOwner.value = true;
    return;
  }
  emit("change-role");
}
</script>

<template>
  <div
    @click.self="$emit('close')"
    class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
    style="pointer-events: auto"
  >
    <div
      class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
      tabindex="-1"
      style="pointer-events: auto"
    >
      <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
        <div v-if="!confirmTransferOwner" class="font-semibold text-lg truncate">
          编辑角色
        </div>
        <div v-if="confirmTransferOwner" class="font-semibold text-lg truncate">
          转移所有权
        </div>
      </header>
      <form v-if="!confirmTransferOwner" @submit.prevent="">
        <p class="mb-4">
          改变 <strong>{{ selectUser.displayName }}</strong> 的角色
        </p>
        <label v-if="canAssignOwner" class="flex mt-3">
          <input
            @click="$emit('set-wantedrole', 'owner')"
            type="radio"
            name="keytype"
            class="radio radio-xs mt-1 mr-1"
            :checked="wantedRole && wantedRole == 'owner'"
          />
          <div class="ml-2">
            <strong class="font-medium block">所有者</strong
            ><span class="text-gray-500 leading-none"
              >拥有该蜃境网络全部权限。蜃境网络中只能有一个所有者。</span
            >
          </div>
        </label>
        <label class="flex mt-3">
          <input
            @click="$emit('set-wantedrole', 'member')"
            type="radio"
            name="keytype"
            class="radio radio-xs mt-1 mr-1"
            :checked="wantedRole && wantedRole == 'member'"
          />
          <div class="ml-2">
            <strong class="font-medium block">普通成员</strong
            ><span class="text-gray-500 leading-none">无法浏览此管理员控制台。</span>
          </div>
        </label>
        <footer class="flex mt-10 justify-end space-x-4">
          <button
            @click="$emit('close')"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            @click="goSubmit"
            :disabled="!wantedRole || wantedRole == selectUser.role"
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
          >
            保存
          </button>
        </footer>
      </form>
      <form v-if="confirmTransferOwner" @submit.prevent="">
        <p class="mb-4">
          更改 <strong>{{ selectUser.displayName }}</strong> 的角色为
          <strong>所有者</strong> 会将
          <strong>{{ selectUser.domainName }}</strong>
          的所有权转移给他。你的角色会被自动更改为 普通成员。
        </p>
        <footer class="flex mt-10 justify-end space-x-4">
          <button
            @click="confirmTransferOwner = false"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            @click="goSubmit"
            :disabled="!wantedRole || wantedRole == selectUser.role"
            class="btn border-0 bg-red-500 hover:bg-red-900 disabled:bg-red-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
          >
            转移所有权
          </button>
        </footer>
      </form>
      <button
        @click="$emit('close')"
        class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
        type="button"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="1.25em"
          height="1.25em"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
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
  --togglehandleborder: 0 0 0 3px white inset,
    var(--handleoffsetcalculator) 0 0 3px white inset;
}

.radio {
  --chkbg: white;
  border-width: 2px;
  border-color: #d6d3d1;
}
.radio:checked {
  --chkbg: white;
  border-width: 5px;
  border-color: #3e5db3;
}
</style>
