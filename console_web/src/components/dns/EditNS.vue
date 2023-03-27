<script setup>
import { ref, watch, computed, onMounted } from "vue";
import { useDisScroll } from "/src/utils.js";

const emit = defineEmits(["nameserver-edited", "close"]);

useDisScroll();

const props = defineProps({
  currentDNS: Object,
  OriDomain: String,
  OriResolver: String,
});

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

const DNSCfgRemovedCurrent = computed(() => {
  var tmpDNSCfg = props.currentDNS;
  if (props.OriDomain == "") {
    if (props.OriResolver == "") return tmpDNSCfg;
    if (!tmpDNSCfg["resolvers"] || tmpDNSCfg["resolvers"].length == 0) {
      var newFallbackResolvers = [];
      for (var i in tmpDNSCfg["fallbackResolvers"]) {
        if (tmpDNSCfg["fallbackResolvers"][i] != props.OriResolver) {
          newFallbackResolvers.push(tmpDNSCfg["fallbackResolvers"][i]);
        }
      }
      tmpDNSCfg["fallbackResolvers"] = newFallbackResolvers;
      return tmpDNSCfg;
    } else {
      var newResolvers = [];
      for (var i in tmpDNSCfg["resolvers"]) {
        if (tmpDNSCfg["resolvers"][i] != props.OriResolver) {
          newResolvers.push(tmpDNSCfg["resolvers"][i]);
        }
      }
      tmpDNSCfg["resolvers"] = newResolvers;
      return tmpDNSCfg;
    }
  } else {
    if (props.OriResolver == "") return tmpDNSCfg;
    var newDomain = [];
    var newRoute = [];
    for (var i in tmpDNSCfg["domains"]) {
      if (tmpDNSCfg["domains"][i] != props.OriDomain) {
        newDomain.push(tmpDNSCfg["domains"][i]);
      } else {
        for (var j in tmpDNSCfg["routes"][props.OriDomain]) {
          if (tmpDNSCfg["routes"][props.OriDomain][j] != props.OriResolver) {
            newRoute.push(tmpDNSCfg["routes"][props.OriDomain][j]);
          }
        }
        tmpDNSCfg["routes"][props.OriDomain] = newRoute;
        if (newRoute.length > 0) {
          newDomain.push(props.OriDomain);
        }
      }
    }
    tmpDNSCfg["domains"] = newDomain;
    if (newRoute.length == 0) {
      var newRoutes = {};
      for (var key in tmpDNSCfg["routes"]) {
        if (key != props.OriDomain) {
          newRoutes[key] = tmpDNSCfg["routes"][key];
        }
      }
      tmpDNSCfg["routes"] = newRoutes;
    }
    return tmpDNSCfg;
  }
});

const invalidNS = ref(false);
const existNS = ref(false);

const isSplitDNS = ref(false);
const nameserverIP = ref("");
const restrictDomain = ref("");
const noSave = computed(() => {
  return (
    nameserverIP.value.length < 6 ||
    (isSplitDNS.value && (!restrictDomain.value || restrictDomain.value.length == 0))
  );
});

watch(
  () => nameserverIP.value,
  (newV) => {
    invalidNS.value = false;
    existNS.value = false;
    nameserverIP.value = nameserverIP.value.toLowerCase().replace(/[^\.:0-9a-f]/gi, "");
  }
);

watch(
  () => restrictDomain.value,
  (newV) => {
    restrictDomain.value = restrictDomain.value.replace(/[^\.:\/'\-_0-9a-zA-Z]/g, "");
  }
);

const ipReg = /((^\s*((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))\s*$)|(^\s*((([0-9a-f]{1,4}:){7}([0-9a-f]{1,4}|:))|(([0-9a-f]{1,4}:){6}(:[0-9a-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9a-f]{1,4}:){5}(((:[0-9a-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9a-f]{1,4}:){4}(((:[0-9a-f]{1,4}){1,3})|((:[0-9a-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9a-f]{1,4}:){3}(((:[0-9a-f]{1,4}){1,4})|((:[0-9a-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9a-f]{1,4}:){2}(((:[0-9a-f]{1,4}){1,5})|((:[0-9a-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9a-f]{1,4}:){1}(((:[0-9a-f]{1,4}){1,6})|((:[0-9a-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9a-f]{1,4}){1,7})|((:[0-9a-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$))/;

function SaveEditedNS() {
  if (
    nameserverIP.value == props.OriResolver &&
    ((restrictDomain.value == props.OriDomain && isSplitDNS.value) ||
      (props.OriDomain == "" && !isSplitDNS.value))
  ) {
    emit("close");
  }
  if (!ipReg.test(nameserverIP.value)) {
    invalidNS.value = true;
    return;
  }
  var reqData = {};
  if (!isSplitDNS.value && props.OriDomain == "") {
    //只修改全局ns值
    reqData = props.currentDNS;
    if (!reqData["resolvers"] || reqData["resolvers"].length == 0) {
      for (var i in reqData["fallbackResolvers"]) {
        if (reqData["fallbackResolvers"][i] == nameserverIP.value) {
          existNS.value = true;
          return;
        }
      }
      for (var i in reqData["fallbackResolvers"]) {
        if (reqData["fallbackResolvers"][i] == props.OriResolver) {
          reqData["fallbackResolvers"][i] = nameserverIP.value;
          break;
        }
      }
    } else {
      for (var i in reqData["resolvers"]) {
        if (reqData["resolvers"][i] == nameserverIP.value) {
          existNS.value = true;
          return;
        }
      }
      for (var i in reqData["resolvers"]) {
        if (reqData["resolvers"][i] == props.OriResolver) {
          reqData["resolvers"][i] = nameserverIP.value;
          break;
        }
      }
    }
  }
  if (!isSplitDNS.value && props.OriDomain != "") {
    //删splitDNS增全局ns
    reqData = DNSCfgRemovedCurrent.value;
    if (!props.currentDNS["resolvers"] || props.currentDNS["resolvers"].length == 0) {
      for (var i in reqData["fallbackResolvers"]) {
        if (reqData["fallbackResolvers"][i] == nameserverIP.value) {
          existNS.value = true;
          return;
        }
      }
      if (!reqData["fallbackResolvers"]) reqData["fallbackResolvers"] = [];
      reqData["fallbackResolvers"].push(nameserverIP.value);
    } else {
      for (var i in reqData["resolvers"]) {
        if (reqData["resolvers"][i] == nameserverIP.value) {
          existNS.value = true;
          return;
        }
      }
      if (!reqData["resolvers"]) reqData["resolvers"] = [];
      reqData["resolvers"].push(nameserverIP.value);
    }
  }
  if (
    isSplitDNS.value &&
    nameserverIP.value != props.OriResolver &&
    restrictDomain.value == props.OriDomain
  ) {
    //splitDNS修改服务器ip地址
    reqData = props.currentDNS;
    for (var j in reqData["routes"][restrictDomain.value]) {
      if (reqData["routes"][restrictDomain.value][j] == nameserverIP.value) {
        existNS.value = true;
        return;
      }
    }
    for (var j in reqData["routes"][restrictDomain.value]) {
      if (reqData["routes"][restrictDomain.value][j] == props.OriResolver) {
        reqData["routes"][restrictDomain.value][j] = nameserverIP.value;
        break;
      }
    }
  }
  if (isSplitDNS.value && restrictDomain.value != props.OriDomain) {
    //删全局NS或原splitDNS，新增splitDNS
    reqData = DNSCfgRemovedCurrent.value;
    var existDomain = false;
    for (var i in reqData["domains"]) {
      if (reqData["domains"][i] == restrictDomain.value) {
        for (var j in reqData["routes"][restrictDomain.value]) {
          if (reqData["routes"][restrictDomain.value][j] == nameserverIP.value) {
            existNS.value = true;
            return;
          }
        }
        if (!reqData["routes"][restrictDomain.value])
          reqData["routes"][restrictDomain.value] = [];
        reqData["routes"][restrictDomain.value].push(nameserverIP.value);
        existDomain = true;
        break;
      }
    }
    if (!existDomain) {
      if (!reqData["domains"]) reqData["domains"] = [];
      reqData["domains"].push(restrictDomain.value);
      reqData["routes"][restrictDomain.value] = [];
      reqData["routes"][restrictDomain.value].push(nameserverIP.value);
    }
  }

  axios
    .post("/admin/api/dns", reqData)
    .then(function (response) {
      if (response.data["status"] == "success") {
        emit("nameserver-edited", response.data["data"]);
      } else {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      toastMsg.value = error;
      toastShow.value = true;
    });
}

onMounted(() => {
  nameserverIP.value = props.OriResolver;
  restrictDomain.value = props.OriDomain;
  if (restrictDomain.value != "") isSplitDNS.value = true;
});
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
        <div class="font-semibold text-lg truncate">编辑域名服务器</div>
      </header>
      <form @submit.prevent="">
        <div class="mt-8 mb-8">
          <label class="font-medium text-gray-900 block mb-1" for="nameserver"
            >域名服务器</label
          >
          <p class="text-sm text-gray-600 mt-2">使用这个IPv4或IPv6地址来解析域名</p>
          <div class="relative mt-2">
            <input
              v-model="nameserverIP"
              :class="{
                'border-stone-200': !invalidNS && !existNS,
                'border-red-400': invalidNS || existNS,
              }"
              class="w-full px-3 z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit tabular-nums"
              id="nameserver"
              type="text"
              placeholder="1.2.3.4"
              autocapitalize="off"
              autocomplete="off"
            />
          </div>
          <p v-if="invalidNS" class="text-sm text-red-400 mt-1">
            不合法，只允许IPv4或IPv6地址
          </p>
          <p v-if="existNS" class="text-sm text-red-400 mt-1">
            该域名解析服务器条目已存在
          </p>
        </div>
        <div class="sm:flex justify-between items-center">
          <div>
            <div class="flex items-center">
              <label class="font-medium text-gray-900" for="restrict">分离 DNS</label
              ><span data-state="closed">
                <div
                  class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 relative ml-2 text-xs"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="0.9em"
                    height="0.9em"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="mr-1"
                  >
                    <polyline points="16 3 21 3 21 8"></polyline>
                    <line x1="4" y1="20" x2="21" y2="3"></line>
                    <polyline points="21 16 21 21 16 21"></polyline>
                    <line x1="15" y1="15" x2="21" y2="21"></line>
                    <line x1="4" y1="4" x2="9" y2="9"></line></svg
                  >分离 DNS
                </div>
              </span>
            </div>
            <p class="text-sm text-gray-600 hidden sm:block">
              此域名服务器将只用于部分域名解析
            </p>
          </div>
          <span data-state="closed">
            <input
              v-model="isSplitDNS"
              id="restrict"
              type="checkbox"
              class="toggle toggle-large mt-2 sm:mt-0 self-start shrink-0"
            />
          </span>
          <p class="text-sm text-gray-600 sm:hidden">此域名服务器将只用于部分域名解析</p>
        </div>
        <div v-if="isSplitDNS" class="mt-4">
          <label class="font-medium text-gray-900 block mb-1" for="searchdomain"
            >检索域</label
          >
          <div class="relative mt-2">
            <input
              v-model="restrictDomain"
              class="w-full px-3 z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
              id="searchdomain"
              type="text"
              placeholder="example.com"
              autocapitalize="off"
              autocomplete="off"
            />
          </div>
          <p class="text-sm text-gray-600 mt-2">
            只用匹配此后缀的单一标签或完全限定查询将使用这个域名服务器
          </p>
        </div>
        <footer class="flex mt-10 justify-end space-x-4">
          <button
            @click="$emit('close')"
            class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            @click="SaveEditedNS"
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
            :disabled="noSave"
          >
            保存
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

  <!-- 提示框显示 -->
  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>
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
</style>
