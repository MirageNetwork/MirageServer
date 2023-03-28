<script setup>
import { watch, ref, onMounted, computed } from "vue";
import { useDisScroll } from "../utils.js";

useDisScroll();

const tenantMenu = ref(null);
const props = defineProps({
  toleft: Number,
  totop: Number,
  selectTenant: Object,
});
const menuLeft = computed(() => {
  return String(String(props.toleft + 32 - tenantMenu.value?.clientWidth));
});
const menuTop = computed(() => {
  if (props.totop <= window.innerHeight / 2) {
    return String(props.totop + 36);
  } else {
    return String(props.totop - 10 - tenantMenu.value?.clientHeight);
  }
});

const emit = defineEmits(["close"]);
const closeMe = (event) => {
  emit("close");
};
</script>

<template>
  <div
    ref="tenantMenu"
    v-click-away="closeMe"
    class="shadow-xl border border-base-300 rounded-md z-20"
    :style="
      'position: fixed; left: ' +
      menuLeft +
      'px; top: ' +
      menuTop +
      'px; min-width: max-content; --radix-popper-transform-origin: 0% 0px;'
    "
  >
    <div
      class="dropdown bg-white rounded-md py-1 z-20"
      style="outline: none; pointer-events: auto"
    >
      <div class="block px-4 py-2 cursor-pointer hover:bg-gray-100">
        共 {{ selectTenant.userCount }} 用户
      </div>
      <div class="block px-4 py-2 cursor-pointer hover:bg-gray-100">
        共 {{ selectTenant.adminCount }} 管理员
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div class="block px-4 py-2 cursor-pointer hover:bg-gray-100">
        共 {{ selectTenant.deviceCount }} 设备
      </div>
      <div class="block px-4 py-2 cursor-pointer hover:bg-gray-100">
        共 {{ selectTenant.subnetCount }} 子网路由
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div class="block px-4 py-2 cursor-pointer hover:bg-gray-100">编辑租户…</div>
      <div class="my-1 border-b border-base-300"></div>
      <div
        @click="$emit('showdialog-removetenant')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 text-red-400"
      >
        移除租户…
      </div>
    </div>
  </div>
</template>

<style scoped></style>
