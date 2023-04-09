<script setup>
import { watch, ref, onMounted, computed } from "vue";
import { useDisScroll } from "../utils.js";

useDisScroll();

const naviMenu = ref(null);
const props = defineProps({
  toleft: Number,
  totop: Number,
  selectNavi: Object,
});
const menuLeft = computed(() => {
  return String(String(props.toleft + 32 - naviMenu.value?.clientWidth));
});
const menuTop = computed(() => {
  if (props.totop <= window.innerHeight / 2) {
    return String(props.totop + 36);
  } else {
    return String(props.totop - 10 - naviMenu.value?.clientHeight);
  }
});

const emit = defineEmits(["close"]);
const closeMe = (event) => {
  emit("close");
};
</script>

<template>
  <div
    ref="naviMenu"
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
      v-if="selectNavi.Arch != 'external'"
      class="dropdown bg-white rounded-md py-1 z-20"
      style="outline: none; pointer-events: auto"
    >
      <div class="block px-2 py-1 hover:bg-gray-100 text-xs">
        版本 {{ selectNavi.Statics.derp.version.split("-")[0] }}
      </div>
      <div class="block px-2 py-1 hover:bg-gray-100 text-xs">
        主机 {{ selectNavi.SSHAddr }}
      </div>
      <div class="block px-2 py-1 hover:bg-gray-100 text-xs">
        接入 {{ selectNavi.Statics.derp.gauge_current_connections }}
      </div>

      <div class="my-1 border-b border-base-300"></div>
      <div class="block px-2 py-1 hover:bg-gray-100 text-xs">
        收 {{ selectNavi.Statics.derp.bytes_received }}B · 发
        {{ selectNavi.Statics.derp.bytes_sent }}B
      </div>
      <div class="block px-2 py-1 hover:bg-gray-100 text-xs">
        收 {{ selectNavi.Statics.derp.packets_received }} 包· 发
        {{ selectNavi.Statics.derp.packets_sent }} 包
      </div>
      <div class="block px-2 py-1 hover:bg-gray-100 text-xs">
        收 {{ selectNavi.Statics.derp.got_ping }} Ping · 发
        {{ selectNavi.Statics.derp.sent_pong }} Pong
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div
        @click="$emit('showdialog-detailinfo')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100"
      >
        更多信息…
      </div>
      <div class="my-1 border-b border-base-300"></div>
      <div
        @click="$emit('showdialog-removenavi')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 text-red-400"
      >
        移除司南…
      </div>
    </div>
    <div
      v-else
      class="dropdown bg-white rounded-md py-1 z-20"
      style="outline: none; pointer-events: auto"
    >
      <div
        @click="$emit('showdialog-removenavi')"
        class="block px-4 py-2 cursor-pointer hover:bg-gray-100 text-red-400"
      >
        移除司南…
      </div>
    </div>
  </div>
</template>

<style scoped></style>
