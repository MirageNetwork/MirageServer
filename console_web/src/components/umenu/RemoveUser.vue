<script setup>
import { ref, watch, computed } from "vue";
import { useDisScroll } from "/src/utils.js";

const emit = defineEmits(["change-role"]);

useDisScroll();

const props = defineProps({
  selectUser: Object,
  userMachineList: Array,
});

const selectrUserRole = computed(() => {
  switch (props.selectUser.role) {
    case "owner":
      return "所有者";
    case "admin":
      return "管理员";
    default:
      return "普通成员";
  }
});

const confirmText = ref("");
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
        <div class="font-semibold text-lg truncate">删除用户</div>
      </header>
      <form @submit.prevent="">
        <p class="mb-4">
          删除
          <strong>{{ selectUser.loginName }} ({{ selectrUserRole }})</strong>
          将会删除他的所有设备。
        </p>
        <h3 class="font-semibold text-gray-800 mb-2">设备列表</h3>
        <div>
          <ul>
            <li v-for="(m, i) in userMachineList" class="flex items-center py-2 border-t">
              <RouterLink
                class="text-gray-600 hover:text-blue-500 truncate"
                :to="'/machines/' + m.addresses[0]"
                target="_blank"
                rel="noreferrer noopener"
                >{{ m.name }}</RouterLink
              >
              <div
                v-if="m.hasSubnets || m.advertisedExitNode"
                class="ml-auto pl-2 tooltip"
                :data-tip="
                  '该设备' + m.advertisedExitNode
                    ? ''
                    : '是一个出口节点' + m.hasSubnets && m.advertisedExitNode
                    ? ''
                    : '且' + m.hasSubnets
                    ? ''
                    : '存在子网转发'
                "
              >
                <span>
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="1.125rem"
                    height="1.125rem"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="text-gray-500"
                  >
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                  </svg>
                </span>
              </div>
            </li>
          </ul>
          <div class="text-gray-500 py-2 items-center border-t">
            <RouterLink
              class="text-blue-500"
              :to="'/machines?q=' + selectUser.loginName"
              target="_blank"
              rel="noopener noreferrer"
              >查看更多信息</RouterLink
            >
          </div>
        </div>
        <p class="mb-2 mt-6">
          输入
          <span class="font-semibold"> {{ selectUser.loginName }}</span> 以确认删除操作：
        </p>
        <div class="flex mb-2">
          <div class="relative w-full z-10">
            <input
              class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
              type="text"
              autocomplete="off"
              autocorrect="off"
              v-model="confirmText"
            />
          </div>
        </div>

        <footer class="flex mt-10 justify-end space-x-4">
          <button
            @click="$emit('close')"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            @click="$emit('confirm-remove')"
            :disabled="confirmText != selectUser.loginName"
            class="btn border-0 bg-red-500 hover:bg-red-900 disabled:bg-red-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
          >
            删除用户
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
