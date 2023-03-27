<script setup>
import { watch, ref, onMounted, computed } from "vue";
import { useDisScroll } from "../utils.js";

useDisScroll();

const devMode = ref(true);

const userMenu = ref(null);
const props = defineProps({
  toleft: Number,
  totop: Number,
  loginName: String,
  cantEdit: Boolean,
});
const menuLeft = computed(() => {
  return String(String(props.toleft + 32 - userMenu.value?.clientWidth));
});
const menuTop = computed(() => {
  if (props.totop <= window.innerHeight / 2) {
    return String(props.totop + 36);
  } else {
    return String(props.totop - 10 - userMenu.value?.clientHeight);
  }
});

const emit = defineEmits(["close"]);
const closeMe = (event) => {
  emit("close");
};
</script>

<template>
  <div
    ref="userMenu"
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
      <RouterLink
        :to="'/machines?q=' + loginName"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
        >查看用户设备
      </RouterLink>
      <RouterLink
        :to="'/logs?actors=[' + loginName + ']'"
        :class="{
          'cursor-pointer hover:bg-gray-100': !devMode,
          'cursor-default text-gray-300': devMode,
        }"
        :disabled="devMode"
        class="block px-4 py-2"
      >
        查看最近活动
      </RouterLink>
      <div class="my-1 border-b border-base-300"></div>
      <div
        @click="$emit('showdialog-changerole')"
        :class="{
          'cursor-pointer hover:bg-gray-100': !cantEdit,
          'cursor-default text-gray-300': cantEdit,
        }"
        :disabled="cantEdit"
        class="block px-4 py-2"
      >
        编辑角色…
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div
        :class="{
          'cursor-pointer hover:bg-gray-100 text-red-400': !cantEdit && !devMode,
          'cursor-default text-gray-300': cantEdit || devMode,
        }"
        :disabled="cantEdit"
        class="block px-4 py-2"
      >
        冻结用户…
      </div>
      <div
        @click="$emit('showdialog-removeuser')"
        :class="{
          'cursor-pointer hover:bg-gray-100 text-red-400': !cantEdit,
          'cursor-default text-gray-300': cantEdit,
        }"
        :disabled="cantEdit"
        class="block px-4 py-2"
      >
        移除用户…
      </div>
    </div>
  </div>
</template>

<style scoped></style>
