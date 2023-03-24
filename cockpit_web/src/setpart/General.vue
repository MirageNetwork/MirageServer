<script setup>
import { onMounted, ref, watch, computed } from "vue";
import Toast from "../components/Toast.vue";

const devMode = ref(true);

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

const MIPv4 = ref("");
const MIPv6 = ref("");
const SrvAddr = ref("");
const ServerURL = ref("");
const BaseDomain = ref("");
const DERPURL = ref("");
const SubnetAccessDueMachine = ref(false);

const isServerURLValid = computed(() => {
  if (ServerURL.value == "") {
    return false;
  }
  const reg = new RegExp("^([\\da-z\\.-]+)\\.([a-z\\.]{2,6})([\\/\\w \\.-]*)*\\/?$");
  return reg.test(ServerURL.value);
});

const setMIPv4Text = ref("设置");
const setMIPv6Text = ref("设置");
const setSrvAddrText = ref("设置");
const setServerURLText = ref("设置");
const setBaseDomainText = ref("设置");
const setDERPURLText = ref("设置");

function setMIPv4() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-mipv4",
      mipv4: MIPv4.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        MIPv4.value = response.data["data"]["mipv4"];
        setMIPv4Text.value = "已设置!";
        setTimeout(() => {
          setMIPv4Text.value = "设置";
        }, 1500);
        toastMsg.value = "已更新MIPv4地址设置";
        toastShow.value = true;
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
function setMIPv6() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-mipv6",
      mipv6: MIPv6.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        MIPv6.value = response.data["data"]["mipv6"];
        setMIPv6Text.value = "已设置!";
        setTimeout(() => {
          setMIPv6Text.value = "设置";
        }, 1500);
        toastMsg.value = "已更新MIPv6地址设置";
        toastShow.value = true;
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
function setSrvAddr() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-srvaddr",
      srvaddr: SrvAddr.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        SrvAddr.value = response.data["data"]["srvaddr"];
        setSrvAddrText.value = "已设置!";
        setTimeout(() => {
          setSrvAddrText.value = "设置";
        }, 1500);
        toastMsg.value = "已更新监听地址设置";
        toastShow.value = true;
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
function setServerURL() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-serverurl",
      ServerURL: ServerURL.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        ServerURL.value = response.data["data"]["server_url"];
        setServerURLText.value = "已设置!";
        setTimeout(() => {
          setServerURLText.value = "设置";
        }, 1500);
        toastMsg.value = "已更新服务域名设置";
        toastShow.value = true;
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
function setBaseDomain() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-basedomain",
      BaseDomain: BaseDomain.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        BaseDomain.value = response.data["data"]["basedomain"];
        setBaseDomainText.value = "已设置!";
        setTimeout(() => {
          setBaseDomainText.value = "设置";
        }, 1500);
        toastMsg.value = "已更新幻域基础域名设置";
        toastShow.value = true;
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
function setDERPURL() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-derpurl",
      DERPURL: DERPURL.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        DERPURL.value = response.data["data"]["derp_url"];
        setDERPURLText.value = "已设置!";
        setTimeout(() => {
          setDERPURLText.value = "设置";
        }, 1500);
        toastMsg.value = "已更新向导节点列表发布地址设置";
        toastShow.value = true;
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
function setSubnetAccessDueMachine() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-routeaccessduemachine",
      SubnetAccessDueMachine: SubnetAccessDueMachine.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        SubnetAccessDueMachine.value = response.data["data"]["route_access_due_machine"];
        if (SubnetAccessDueMachine.value) {
          toastMsg.value = "已启用子网路由访问控制依据节点设置";
        } else {
          toastMsg.value = "已禁用子网路由访问控制依据节点设置";
        }
        toastShow.value = true;
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
  axios
    .get("/cockpit/api/setting/general")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        ServerURL.value = response.data["data"]["server_url"];
        SrvAddr.value = response.data["data"]["srvaddr"];
        MIPv4.value = response.data["data"]["mipv4"];
        MIPv6.value = response.data["data"]["mipv6"];
        BaseDomain.value = response.data["data"]["basedomain"];
        DERPURL.value = response.data["data"]["derp_url"];
        SubnetAccessDueMachine.value = response.data["data"]["route_access_due_machine"];
      } else {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      alert(error);
    });
});
</script>

<template>
  <div class="flex-1">
    <div
      class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center"
    >
      <h1 class="mr-2" tabindex="-1">基本配置</h1>
    </div>
    <div class="text-gray-600 mt-3">
      <p>此页为蜃境控制器启动所必须的基本配置信息，请务必确保正确性</p>
    </div>
    <div class="mt-6 space-y-6">
      <div>
        <header class="max-w-2xl">
          <h3 class="text-xl font-semibold tracking-tight">控制器服务监听</h3>
        </header>
        <p class="mt-3 text-gray-600">设置控制器（及控制台）服务监听地址</p>
        <p class="text-gray-400 text-sm">
          默认为 <code class="bg-gray-200 text-xs rounded px-1">:8080</code>
        </p>
        <div
          class="mt-1 max-w-sm flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="SrvAddr"
          />
          <button
            @click="setSrvAddr"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
          >
            {{ setSrvAddrText }}
          </button>
        </div>
        <p class="mt-3 text-gray-600">设置控制器（及控制台）服务域名</p>
        <p class="text-gray-400 text-sm">
          参考形式
          <code class="bg-gray-200 text-xs rounded px-1">sdp.mirage.com</code
          >，请勿添加协议头并确保通过前置代理设置了HTTPS
        </p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': !isServerURLValid,
            'border-stone-200 hover:border-stone-400': isServerURLValid,
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="ServerURL"
          />
          <button
            :disabled="!isServerURLValid"
            @click="setServerURL"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 disabled:text-gray-200 font-medium text-sm whitespace-nowrap"
          >
            {{ setServerURLText }}
          </button>
        </div>
      </div>
      <div>
        <header class="max-w-2xl">
          <h3 class="text-xl font-semibold tracking-tight">蜃境IP网段</h3>
        </header>
        <p class="mt-3 text-gray-600">这是全部蜃境节点的IPv4网段</p>
        <p class="text-gray-400 text-sm">
          CIDR格式，默认为
          <code class="bg-gray-200 text-xs rounded px-1">100.64.0.0/10</code>
        </p>
        <div
          class="mt-1 max-w-sm flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="MIPv4"
          />
          <button
            @click="setMIPv4"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
          >
            {{ setMIPv4Text }}
          </button>
        </div>
        <p class="mt-3 text-gray-600">这是全部蜃境节点的IPv6网段</p>
        <p class="text-gray-400 text-sm">
          CIDR格式，默认为
          <code class="bg-gray-200 text-xs rounded px-1">fd7a:115c:a1e0::/48</code>
        </p>
        <div
          class="mt-1 max-w-sm flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="MIPv6"
          />
          <button
            @click="setMIPv6"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
          >
            {{ setMIPv6Text }}
          </button>
        </div>
      </div>
      <div>
        <header class="max-w-2xl">
          <h3 class="text-xl font-semibold tracking-tight">幻域基础域</h3>
        </header>
        <p class="mt-3 text-gray-600">这是用于拼接幻域域名的基础域</p>
        <p class="text-gray-400 text-sm">
          默认值
          <code class="bg-gray-200 text-xs rounded px-1">mira.net</code
          >，启用幻域后，将可通过形式如
          <code class="bg-gray-200 text-xs rounded px-1">win11.orgcode.mira.net</code>
          的域名访问节点
        </p>
        <div
          class="mt-1 max-w-sm flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="BaseDomain"
          />
          <button
            @click="setBaseDomain"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
          >
            {{ setBaseDomainText }}
          </button>
        </div>
      </div>
      <div>
        <header class="max-w-2xl">
          <h3 class="text-xl font-semibold tracking-tight">向导节点列表发布地址</h3>
        </header>
        <p class="mt-3 text-gray-600">从该地址获取向导节点列表</p>
        <p class="text-gray-400 text-sm">
          默认值
          <code class="bg-gray-200 text-xs rounded px-1"
            >https://controlplane.tailscale.com/derpmap/default</code
          >
        </p>
        <div
          class="mt-1 max-w-sm flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="DERPURL"
          />
          <button
            @click="setDERPURL"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
          >
            {{ setDERPURLText }}
          </button>
        </div>
      </div>
      <div>
        <header class="max-w-2xl">
          <h3 class="text-xl font-semibold tracking-tight">ACL特性（临时）</h3>
        </header>
        <div class="flex flex-row items-center">
          <input
            v-model="SubnetAccessDueMachine"
            @change="setSubnetAccessDueMachine"
            type="checkbox"
            class="toggle"
          />
          <div class="ml-4 flex flex-col">
            <p class="text-gray-600">
              是否依据节点的可访问性自动设定节点所转发子网的可访问性
            </p>
            <p class="text-gray-400 text-sm">该项目后续要移到租户设置中</p>
          </div>
        </div>
      </div>
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

.tooltip {
  --tooltip-color: #faf9f8;
  --tooltip-text-color: #3a3939;
  text-align: start;
  white-space: normal;
}

.tooltip:before {
  max-width: 16rem;
  font-size: small;
  font-weight: 300;
  border-radius: 0.375rem;
  box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  padding-left: 0.75rem;
  padding-right: 0.75rem;
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  border-width: 1px;
  border-color: #e1dfde;
}
</style>
