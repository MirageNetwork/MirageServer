<script setup>
import { watch, ref, onMounted, computed } from "vue";
import { useDisScroll } from "../utils.js";

useDisScroll();

const machineMenu = ref(null);
const props = defineProps({
  toleft: Number,
  totop: Number,
  neverExpires: Boolean,
});
const menuLeft = computed(() => {
  return String(String(props.toleft + 32 - machineMenu.value?.clientWidth));
});
const menuTop = computed(() => {
  if (props.totop <= window.innerHeight / 2) {
    return String(props.totop + 36);
  } else {
    return String(props.totop - 10 - machineMenu.value?.clientHeight);
  }
});

const emit = defineEmits(["close"]);
const closeMe = (event) => {
  emit("close");
};
</script>

<template>
  <div
    ref="machineMenu"
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
      style="
        outline: none;
        --radix-dropdown-menu-content-transform-origin: var(
          --radix-popper-transform-origin
        );
        pointer-events: auto;
      "
    >
      <div
        @click="$emit('showdialog-updatehostname')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
      >
        编辑设备名称…
      </div>
      <div
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
      >
        分享…
      </div>
      <div
        @click="$emit('set-expires')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
      >
        {{ neverExpires ? "启用密钥过期" : "禁用密钥过期" }}
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div
        @click="$emit('showdialog-setsubnet')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
      >
        编辑子网转发…
      </div>
      <div
        @click="$emit('showdialog-edittags')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
      >
        编辑ACL标签…
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div
        @click="$emit('showdialog-remove')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100 text-red-400"
      >
        移除…
      </div>
    </div>
  </div>
</template>

<style scoped></style>
