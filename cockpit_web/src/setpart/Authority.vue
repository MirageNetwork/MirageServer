<script setup>
import { onMounted, ref, watch, computed } from "vue";
import Toast from "../components/Toast.vue";
import MicrosoftHelp from "./help/Microsoft.vue";
import GithubHelp from "./help/Github.vue";
import GoogleHelp from "./help/Google.vue";
import AppleHelp from "./help/Apple.vue";

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

const microsoftHelpShow = ref(false);
const githubHelpShow = ref(false);
const googleHelpShow = ref(false);
const appleHelpShow = ref(false);

const ESURL = ref("");
const ESKey = ref("");
const WXScanURL = ref("");

const SMS = ref({});
const IDaaS = ref({});
const OIDC = ref({});
const Microsoft = ref({});
const Github = ref({});
const Google = ref({});
const Apple = ref({});

const enableAli = ref(false);

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
function setMicrosoft() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-microsoft",
      Microsoft: Microsoft.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        Microsoft.value = response.data["data"]["microsoft"];
        toastMsg.value = "已更新Microsoft设置";
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
function setGithub() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-github",
      Github: Github.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        Github.value = response.data["data"]["github"];
        toastMsg.value = "已更新Github设置";
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
function setGoogle() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-google",
      Google: Google.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        Google.value = response.data["data"]["google"];
        toastMsg.value = "已更新Google设置";
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
function setApple() {
  axios
    .post("/cockpit/api/setting/general", {
      state: "set-apple",
      Apple: Apple.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        Apple.value = response.data["data"]["apple"];
        toastMsg.value = "已更新Apple设置";
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
        Microsoft.value = response.data["data"]["microsoft"];
        Github.value = response.data["data"]["github"];
        Google.value = response.data["data"]["google"];
        Apple.value = response.data["data"]["apple"];
      } else {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      toastMsg.value = error;
      toastShow.value = true;
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
      <p>此页为蜃境控制器服务所需的第三方服务信息，包括日志服务、身份鉴别服务等。</p>
    </div>
    <div class="mt-6 space-y-6">
      <!---->
      <div>
        <header class="max-w-sm flex">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="48"
            height="28"
            preserveAspectRatio="xMinYMin meet"
            viewBox="0 0 256 256"
            id="elasticsearch"
          >
            <path
              fill="#FFF"
              d="M255.96 134.393c0-21.521-13.373-40.117-33.223-47.43a75.239 75.239 0 0 0 1.253-13.791c0-39.909-32.386-72.295-72.295-72.295-23.193 0-44.923 11.074-58.505 30.088-6.686-5.224-14.835-7.94-23.402-7.94-21.104 0-38.446 17.133-38.446 38.446 0 4.597.836 9.194 2.298 13.373C13.582 81.739 0 100.962 0 122.274c0 21.522 13.373 40.327 33.431 47.64-.835 4.388-1.253 8.985-1.253 13.79 0 39.7 32.386 72.087 72.086 72.087 23.402 0 44.924-11.283 58.505-30.088 6.686 5.223 15.044 8.149 23.611 8.149 21.104 0 38.446-17.134 38.446-38.446 0-4.597-.836-9.194-2.298-13.373 19.64-7.104 33.431-26.327 33.431-47.64z"
            ></path>
            <path
              fill="#F4BD19"
              d="M100.085 110.364l57.043 26.119 57.669-50.565a64.312 64.312 0 0 0 1.253-12.746c0-35.52-28.834-64.355-64.355-64.355-21.313 0-41.162 10.447-53.072 27.998l-9.612 49.73 11.074 23.82z"
            ></path>
            <path
              fill="#3CBEB1"
              d="M40.953 170.75c-.835 4.179-1.253 8.567-1.253 12.955 0 35.52 29.043 64.564 64.564 64.564 21.522 0 41.372-10.656 53.49-28.208l9.403-49.729-12.746-24.238-57.251-26.118-56.207 50.774z"
            ></path>
            <path
              fill="#E9478C"
              d="M40.536 71.918l39.073 9.194 8.775-44.506c-5.432-4.179-11.91-6.268-18.805-6.268-16.925 0-30.924 13.79-30.924 30.924 0 3.552.627 7.313 1.88 10.656z"
            ></path>
            <path
              fill="#2C458F"
              d="M37.192 81.32c-17.551 5.642-29.67 22.567-29.67 40.954 0 17.97 11.074 34.059 27.79 40.327l54.953-49.73-10.03-21.52-43.043-10.03z"
            ></path>
            <path
              fill="#95C63D"
              d="M167.784 219.852c5.432 4.18 11.91 6.478 18.596 6.478 16.925 0 30.924-13.79 30.924-30.924 0-3.761-.627-7.314-1.88-10.657l-39.073-9.193-8.567 44.296z"
            ></path>
            <path
              fill="#176655"
              d="M175.724 165.317l43.043 10.03c17.551-5.85 29.67-22.566 29.67-40.954 0-17.97-11.074-33.849-27.79-40.326l-56.415 49.311 11.492 21.94z"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
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
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            width="28"
            height="28"
          >
            <path fill="none" d="M0 0h24v24H0z" />
            <path
              d="M15.84 12.691l-.067.02a1.522 1.522 0 0 1-.414.062c-.61 0-.954-.412-.77-.921.136-.372.491-.686.925-.831.672-.245 1.142-.804 1.142-1.455 0-.877-.853-1.587-1.905-1.587s-1.904.71-1.904 1.587v4.868c0 1.17-.679 2.197-1.694 2.778a3.829 3.829 0 0 1-1.904.502c-1.984 0-3.598-1.471-3.598-3.28 0-.576.164-1.117.451-1.587.444-.73 1.184-1.287 2.07-1.541a1.55 1.55 0 0 1 .46-.073c.612 0 .958.414.773.924-.126.347-.466.645-.861.803a2.162 2.162 0 0 0-.139.052c-.628.26-1.061.798-1.061 1.422 0 .877.853 1.587 1.905 1.587s1.904-.71 1.904-1.587V9.566c0-1.17.679-2.197 1.694-2.778a3.829 3.829 0 0 1 1.904-.502c1.984 0 3.598 1.471 3.598 3.28 0 .576-.164 1.117-.451 1.587-.442.726-1.178 1.282-2.058 1.538zM2 12c0 5.523 4.477 10 10 10s10-4.477 10-10S17.523 2 12 2 2 6.477 2 12z"
              fill="rgba(56,186,109,1)"
            />
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4">
            微信小程序扫码登录服务
          </h3>
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
      <div v-if="enableAli">
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
      <div v-if="enableAli">
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
      <div v-if="enableAli">
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
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex mt-4">
          <svg
            t="1679382174674"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="2682"
            width="112"
            height="28"
          >
            <path d="M0 0h486.592v486.592H0z" fill="#F25022" p-id="2683"></path>
            <path d="M537.408 0H1024v486.592H537.408z" fill="#7FBA00" p-id="2684"></path>
            <path d="M0 537.408h486.592V1024H0z" fill="#00A4EF" p-id="2685"></path>
            <path
              d="M537.408 537.408H1024V1024H537.408z"
              fill="#FFB900"
              p-id="2686"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
            Microsoft 认证设置
          </h3>
          <div
            @click="microsoftHelpShow = true"
            class="tooltip flex items-center text-stone-300"
            data-tip="点击查看帮助"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1em"
              height="1em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.35"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="ml-1"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
          </div>
          <div class="w-full flex justify-end">
            <button
              :disabled="Microsoft.client_id == '' || Microsoft.client_secret == ''"
              @click="setMicrosoft"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Client ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Microsoft.client_id == '',
            'border-stone-200 hover:border-stone-400': Microsoft.client_id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Microsoft.client_id"
          />
        </div>
        <p class="mt-3 text-gray-600">Client Secret</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Microsoft.client_secret == '',
            'border-stone-200 hover:border-stone-400': Microsoft.client_secret != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Microsoft.client_secret"
          />
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex mt-4">
          <svg
            t="1679387527759"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="3364"
            width="112"
            height="28"
          >
            <path
              d="M0 524.714667c0 223.36 143.146667 413.269333 342.656 482.986666 26.88 6.826667 22.784-12.373333 22.784-25.344v-88.618666c-155.136 18.176-161.322667-84.48-171.818667-101.589334-21.077333-35.968-70.741333-45.141333-55.936-62.250666 35.328-18.176 71.338667 4.608 112.981334 66.261333 30.165333 44.672 89.002667 37.12 118.912 29.653333a144.64 144.64 0 0 1 39.68-69.546666c-160.682667-28.757333-227.712-126.848-227.712-243.541334 0-56.576 18.688-108.586667 55.253333-150.570666-23.296-69.205333 2.176-128.384 5.546667-137.173334 66.474667-5.973333 135.424 47.573333 140.8 51.754667 37.76-10.197333 80.810667-15.573333 128.981333-15.573333 48.426667 0 91.733333 5.546667 129.706667 15.872 12.8-9.813333 76.885333-55.765333 138.666666-50.133334 3.285333 8.789333 28.16 66.602667 6.272 134.826667 37.077333 42.069333 55.936 94.549333 55.936 151.296 0 116.864-67.413333 215.04-228.565333 243.456a145.92 145.92 0 0 1 43.52 104.106667v128.64c0.896 10.282667 0 20.48 17.194667 20.48 202.410667-68.224 348.16-259.541333 348.16-484.906667C1023.018667 242.176 793.941333 13.312 511.573333 13.312 228.864 13.184 0 242.090667 0 524.714667z"
              fill="#000000"
              p-id="3365"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
            GitHub 认证设置
          </h3>
          <div
            @click="githubHelpShow = true"
            class="tooltip flex items-center text-stone-300"
            data-tip="点击查看帮助"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1em"
              height="1em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.35"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="ml-1"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
          </div>
          <div class="w-full flex justify-end">
            <button
              :disabled="Github.client_id == '' || Github.client_secret == ''"
              @click="setGithub"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Client ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Github.client_id == '',
            'border-stone-200 hover:border-stone-400': Github.client_id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Github.client_id"
          />
        </div>
        <p class="mt-3 text-gray-600">Client Secret</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Github.client_secret == '',
            'border-stone-200 hover:border-stone-400': Github.client_secret != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Github.client_secret"
          />
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex mt-4">
          <svg
            t="1679449475826"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="4669"
            width="112"
            height="28"
          >
            <path
              d="M214.101333 512c0-32.512 5.546667-63.701333 15.36-92.928L57.173333 290.218667A491.861333 491.861333 0 0 0 4.693333 512c0 79.701333 18.858667 154.88 52.394667 221.610667l172.202667-129.066667A290.56 290.56 0 0 1 214.101333 512"
              fill="#FBBC05"
              p-id="4670"
            ></path>
            <path
              d="M516.693333 216.192c72.106667 0 137.258667 25.002667 188.458667 65.962667L854.101333 136.533333C763.349333 59.178667 646.997333 11.392 516.693333 11.392c-202.325333 0-376.234667 113.28-459.52 278.826667l172.373334 128.853333c39.68-118.016 152.832-202.88 287.146666-202.88"
              fill="#EA4335"
              p-id="4671"
            ></path>
            <path
              d="M516.693333 807.808c-134.357333 0-247.509333-84.864-287.232-202.88l-172.288 128.853333c83.242667 165.546667 257.152 278.826667 459.52 278.826667 124.842667 0 244.053333-43.392 333.568-124.757333l-163.584-123.818667c-46.122667 28.458667-104.234667 43.776-170.026666 43.776"
              fill="#34A853"
              p-id="4672"
            ></path>
            <path
              d="M1005.397333 512c0-29.568-4.693333-61.44-11.648-91.008H516.650667V614.4h274.602666c-13.696 65.962667-51.072 116.650667-104.533333 149.632l163.541333 123.818667c93.994667-85.418667 155.136-212.650667 155.136-375.850667"
              fill="#4285F4"
              p-id="4673"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
            Google 认证设置
          </h3>
          <div
            @click="googleHelpShow = true"
            class="tooltip flex items-center text-stone-300"
            data-tip="点击查看帮助"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1em"
              height="1em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.35"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="ml-1"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
          </div>
          <div class="w-full flex justify-end">
            <button
              :disabled="Google.client_id == '' || Google.client_secret == ''"
              @click="setGoogle"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Client ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Google.client_id == '',
            'border-stone-200 hover:border-stone-400': Google.client_id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Google.client_id"
          />
        </div>
        <p class="mt-3 text-gray-600">Client Secret</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Google.client_secret == '',
            'border-stone-200 hover:border-stone-400': Google.client_secret != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Google.client_secret"
          />
        </div>
      </div>
      <!---->
      <div>
        <header class="max-w-sm flex mt-4">
          <svg
            t="1679468518353"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="1724"
            width="112"
            height="28"
          >
            <path
              d="M645.289723 165.758826C677.460161 122.793797 701.865322 62.036894 693.033384 0c-52.607627 3.797306-114.089859 38.61306-149.972271 84.010072-32.682435 41.130375-59.562245 102.313942-49.066319 161.705521 57.514259 1.834654 116.863172-33.834427 151.294929-79.956767zM938.663644 753.402663c-23.295835 52.820959-34.517089 76.415459-64.511543 123.177795-41.855704 65.279538-100.905952 146.644295-174.121433 147.198957-64.980873 0.725328-81.748754-43.30636-169.982796-42.751697-88.234042 0.46933-106.623245 43.605024-171.732117 42.965029-73.130149-0.682662-129.065752-74.026142-170.964122-139.348347-117.11917-182.441374-129.44975-396.626524-57.172928-510.545717 51.327636-80.895427 132.393729-128.212425 208.553189-128.212425 77.482118 0 126.207106 43.519692 190.377318 43.519692 62.292892 0 100.137957-43.647691 189.779989-43.647691 67.839519 0 139.732344 37.802399 190.889315 103.03927-167.678812 94.036667-140.543004 339.069598 28.885128 404.605134z"
              fill="#0B0B0A"
              p-id="1725"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
            Apple 认证设置
          </h3>
          <div
            @click="appleHelpShow = true"
            class="tooltip flex items-center text-stone-300"
            data-tip="点击查看帮助"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1em"
              height="1em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.35"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="ml-1"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
          </div>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                Apple.client_id == '' ||
                Apple.team_id == '' ||
                Apple.key_id == '' ||
                Apple.private_key == ''
              "
              @click="setApple"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              保存
            </button>
          </div>
        </header>
        <p class="mt-3 text-gray-600">Client ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Apple.client_id == '',
            'border-stone-200 hover:border-stone-400': Apple.client_id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Apple.client_id"
          />
        </div>
        <p class="mt-3 text-gray-600">Team ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Apple.team_id == '',
            'border-stone-200 hover:border-stone-400': Apple.team_id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Apple.team_id"
          />
        </div>
        <p class="mt-3 text-gray-600">Key ID</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Apple.key_id == '',
            'border-stone-200 hover:border-stone-400': Apple.key_id != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="Apple.key_id"
          />
        </div>
        <p class="mt-3 text-gray-600">Private Key</p>
        <div
          :class="{
            'border-red-500 hover:border-red-700': Apple.private_key == '',
            'border-stone-200 hover:border-stone-400': Apple.private_key != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <textarea
            class="textarea outline-none py-2 px-3 w-full h-52 font-mono text-sm text-ellipsis"
            v-model="Apple.private_key"
          />
        </div>
      </div>
      <!---->
    </div>
  </div>

  <!-- 菜单弹出提示框显示 -->
  <Teleport to="body">
    <MicrosoftHelp
      v-if="microsoftHelpShow"
      @close="microsoftHelpShow = false"
    ></MicrosoftHelp>
    <GithubHelp v-if="githubHelpShow" @close="githubHelpShow = false"></GithubHelp>
    <GoogleHelp v-if="googleHelpShow" @close="googleHelpShow = false"></GoogleHelp>
    <AppleHelp v-if="appleHelpShow" @close="appleHelpShow = false"></AppleHelp>
  </Teleport>
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
  font-size: x-small;
  font-weight: 300;
  border-radius: 0.375rem;
  box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  padding-left: 0.25rem;
  padding-right: 0.25rem;
  padding-top: 0rem;
  padding-bottom: 0rem;
  border-width: 1px;
  border-color: #e1dfde;
}
</style>
