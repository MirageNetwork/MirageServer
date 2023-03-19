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

const ESURL = ref("");
const ESKey = ref("");
const WXScanURL = ref("");

const SMS = ref({});
const IDaaS = ref({});
const OIDC = ref({});

const newScope = ref("");
const newExtra = ref({});

function isValidURL(text) {
  if (text == "") {
    return false;
  }
  const reg = new RegExp(
    "^(https?:\\/\\/)?" + // protocol
      "((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|" + // domain name
      "((\\d{1,3}\\.){3}\\d{1,3}))" + // OR ip (v4) address
      "(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*" + // port and path
      "(\\?[;&a-z\\d%_.~+=-]*)?" + // query string
      "(\\#[-a-z\\d_]*)?$", // fragment locator
    "i"
  );
  return reg.test(text);
}

const setWXScanURLText = ref("设置");

function setES() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-es",
      ESURL: ESURL.value,
      ESKey: ESKey.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        ESKey.value = response.data["data"]["es_key"];
        ESURL.value = response.data["data"]["es_url"];
        toastMsg.value = "已更新日志服务API设置";
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
function setWXScanURL() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-wxscanurl",
      WXScanURL: WXScanURL.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        WXScanURL.value = response.data["data"]["wxscan_url"];
        setWXScanURLText.value = "已设置!";
        setTimeout(() => {
          setWXScanURLText.value = "设置";
        }, 1500);
        toastMsg.value = "已更新微信小程序扫码登录服务URL设置";
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
function setSMS() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-sms",
      SMS: SMS.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        SMS.value = response.data["data"]["sms"];
        toastMsg.value = "已更新短信设置";
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
function setIDaaS() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-idaas",
      IDaaS: IDaaS.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        IDaaS.value = response.data["data"]["idaas"];
        toastMsg.value = "已更新IDaaS管理设置";
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

function addScope() {
  let duplicateScopes = [];
  newScope.value.split(",").forEach((scope) => {
    if (OIDC.value["scope"].indexOf(scope.trim()) != -1) {
      duplicateScopes.push(scope.trim());
    } else {
      if (scope.trim() != "") {
        OIDC.value["scope"].push(scope.trim());
      }
    }
  });
  toastMsg.value = "已存在Scope: " + duplicateScopes.join(",");
  toastShow.value = true;
  newScope.value = "";
}
function rmScope(scope) {
  OIDC.value["scope"].splice(OIDC.value["scope"].indexOf(scope), 1);
}

function addExtra() {
  if (
    OIDC.value["extra"][newExtra.value["k"]] &&
    OIDC.value["extra"][newExtra.value["k"]] != ""
  ) {
    console.log(OIDC.value["extra"][newExtra.value["k"]]);
    toastMsg.value = "已存在该Extra";
    toastShow.value = true;
    return;
  }
  OIDC.value["extra"][newExtra.value["k"]] = newExtra.value["v"];
  newExtra.value["k"] = "";
  newExtra.value["v"] = "";
}
function rmExtra(k) {
  delete OIDC.value["extra"][k];
}

function setOIDC() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-oidc",
      OIDC: OIDC.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        OIDC.value = response.data["data"]["oidc"];
        toastMsg.value = "已更新OIDC设置";
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
        ESURL.value = response.data["data"]["es_url"];
        ESKey.value = response.data["data"]["es_key"];
        WXScanURL.value = response.data["data"]["wxscan_url"];
        SMS.value = response.data["data"]["sms"];
        IDaaS.value = response.data["data"]["idaas"];
        OIDC.value = response.data["data"]["oidc"];
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
      <h1 class="mr-2" tabindex="-1">第三方服务</h1>
    </div>
    <div class="text-gray-600 mt-3">
      <p>
        此页为蜃境控制器服务所需的第三方服务信息，包括日志服务、短信服务、身份鉴别服务等。
      </p>
      <p class="text-gray-400 text-sm">在蜃境控制台完成对应适配前，均为必填</p>
    </div>
    <div class="mt-6 space-y-6">
      <!---->
      <div>
        <header class="max-w-sm flex">
          <h3 class="text-xl font-semibold tracking-tight mr-4 min-w-fit">
            日志服务配置
          </h3>
          <div class="w-full flex justify-end">
            <button
              :disabled="!isValidURL(ESURL) || ESKey == ''"
              @click="setES"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Elasticsearch URL</p>
        <p class="text-gray-400 text-sm">
          参考形式
          <code class="bg-gray-200 text-xs rounded px-1">https://172.17.0.1:9200</code>
        </p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': !isValidURL(ESURL),
            'border-stone-200 hover:border-stone-400': isValidURL(ESURL),
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="ESURL"
          />
        </div>
        <p class="mt-3 text-gray-600">Elasticsearch API Key</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': ESKey == '',
            'border-stone-200 hover:border-stone-400': ESKey != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="ESKey"
          />
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex">
          <h3 class="text-xl font-semibold tracking-tight">微信小程序扫码登录服务</h3>
        </header>
        <p class="mt-3 text-gray-600">扫码小程序服务URL</p>
        <p class="text-gray-400 text-sm">
          参考形式
          <code class="bg-gray-200 text-xs rounded px-1">https://wxscan.mirage.com</code>
        </p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': !isValidURL(WXScanURL),
            'border-stone-200 hover:border-stone-400': isValidURL(WXScanURL),
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="WXScanURL"
          />
          <button
            :disabled="!isValidURL(WXScanURL)"
            @click="setWXScanURL"
            class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 disabled:text-gray-200 font-medium text-sm whitespace-nowrap"
          >
            {{ setWXScanURLText }}
          </button>
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex">
          <h3 class="text-xl font-semibold tracking-tight mr-4 min-w-fit">
            短信通知服务（阿里云）
          </h3>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                SMS.id == '' || SMS.key == '' || SMS.sign == '' || SMS.template == ''
              "
              @click="setSMS"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Access ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': SMS.id == '',
            'border-stone-200 hover:border-stone-400': SMS.id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="SMS.id"
          />
        </div>
        <p class="mt-3 text-gray-600">Access Key</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': SMS.key == '',
            'border-stone-200 hover:border-stone-400': SMS.key != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="SMS.key"
          />
        </div>
        <p class="mt-3 text-gray-600">短信签名</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': SMS.sign == '',
            'border-stone-200 hover:border-stone-400': SMS.sign != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="SMS.sign"
          />
        </div>
        <p class="mt-3 text-gray-600">短信模板</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': SMS.template == '',
            'border-stone-200 hover:border-stone-400': SMS.template != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="SMS.template"
          />
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex">
          <h3 class="text-xl font-semibold tracking-tight mr-4 min-w-fit">
            阿里云IDaaS用户管理API
          </h3>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                IDaaS.app == '' ||
                IDaaS.id == '' ||
                IDaaS.key == '' ||
                IDaaS.instance == '' ||
                IDaaS.org == ''
              "
              @click="setIDaaS"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">App ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': IDaaS.app == '',
            'border-stone-200 hover:border-stone-400': IDaaS.app != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="IDaaS.app"
          />
        </div>
        <p class="mt-3 text-gray-600">Client ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': IDaaS.id == '',
            'border-stone-200 hover:border-stone-400': IDaaS.id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="IDaaS.id"
          />
        </div>
        <p class="mt-3 text-gray-600">Client Key</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': IDaaS.key == '',
            'border-stone-200 hover:border-stone-400': IDaaS.key != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="IDaaS.key"
          />
        </div>
        <p class="mt-3 text-gray-600">实例</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': IDaaS.instance == '',
            'border-stone-200 hover:border-stone-400': IDaaS.instance != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="IDaaS.instance"
          />
        </div>
        <p class="mt-3 text-gray-600">组织ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': IDaaS.org == '',
            'border-stone-200 hover:border-stone-400': IDaaS.org != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="IDaaS.org"
          />
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex">
          <h3 class="text-xl font-semibold tracking-tight mr-4 min-w-fit">
            阿里云IDaaS OIDC服务设置
          </h3>
          <div class="w-full flex justify-end">
            <button
              :disabled="!isValidURL(OIDC.issuer) || OIDC.id == '' || OIDC.key == ''"
              @click="setOIDC"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Issuer</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': !isValidURL(OIDC.issuer),
            'border-stone-200 hover:border-stone-400': isValidURL(OIDC.issuer),
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="OIDC.issuer"
          />
        </div>
        <p class="mt-3 text-gray-600">Client ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': OIDC.id == '',
            'border-stone-200 hover:border-stone-400': OIDC.id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="OIDC.id"
          />
        </div>
        <p class="mt-3 text-gray-600">Client Key</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': OIDC.key == '',
            'border-stone-200 hover:border-stone-400': OIDC.key != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="OIDC.key"
          />
        </div>
        <div class="mt-3 flex flex-row items-center max-w-sm">
          <p class="text-gray-600 mr-4">Scope</p>
          <div
            class="w-full h-8 items-center flex border-stone-200 hover:border-stone-400 border rounded-md relative overflow-hidden min-w-0"
          >
            <input
              class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
              v-model="newScope"
            />
            <button
              :disabled="newScope.trim() == ''"
              @click="addScope"
              class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 disabled:text-gray-200 font-medium text-xl whitespace-nowrap"
            >
              +
            </button>
          </div>
        </div>
        <p class="text-gray-400 text-xs">英文逗号分隔可一次添加多个</p>
        <div class="flex flex-wrap gap-2 p-1 mt-2 max-w-sm">
          <span v-for="(scopeItem, i) in OIDC.scope">
            <div
              class="flex items-center align-middle justify-center font-medium border-stone-400 bg-stone-200 border rounded-full px-2 py-1 leading-none text-xs"
            >
              <span class="text-gray-500">{{ scopeItem }}</span>
              <span class="ml-1">
                <button @click="rmScope(scopeItem)" type="button">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="12"
                    height="12"
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
              </span>
            </div>
          </span>
        </div>
        <div class="mt-3 flex flex-row items-center max-w-sm">
          <p class="text-gray-600 mr-4 min-w-fit">额外参数</p>
          <div
            class="w-full h-8 flex items-center border-stone-200 hover:border-stone-400 border rounded-md relative overflow-hidden min-w-0"
          >
            <input
              class="border-r bg-stone-50 outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
              v-model="newExtra.k"
            />
            <input
              class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
              v-model="newExtra.v"
            />
            <button
              :disabled="
                !newExtra.k ||
                !newExtra.v ||
                newExtra.k.trim() == '' ||
                newExtra.v.trim() == ''
              "
              @click="addExtra"
              class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 disabled:text-gray-200 font-medium text-xl whitespace-nowrap"
            >
              +
            </button>
          </div>
        </div>
        <div class="flex flex-wrap gap-2 p-1 mt-2 max-w-sm">
          <span v-for="(v, k) in OIDC.extra">
            <div
              class="flex items-center align-middle justify-center font-medium border-stone-400 bg-stone-50 border rounded-full leading-none text-xs"
            >
              <span
                class="pl-2 pr-1 py-1 text-gray-500 bg-stone-200 rounded-l-full h-full"
                >{{ k }}</span
              >
              <span class="pl-1 text-gray-500">{{ v }}</span>
              <span class="ml-1 pr-2">
                <button @click="rmExtra(k)" type="button">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="12"
                    height="12"
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
              </span>
            </div>
          </span>
        </div>
        <!---->
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
