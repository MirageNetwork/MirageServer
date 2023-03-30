<script setup>
import { onMounted, ref, watch, computed } from "vue";
import { genFileId } from "element-plus";
import Toast from "../components/Toast.vue";

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

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

const ClientVersion = ref({});

const winUploadURL = ref("");
const wantWinVersion = ref({});
const winExtURL = ref(false);
const winUploader = ref(null);
const winFileName = ref("");
const winProc = ref("");
const winProcPercent = ref(0);

function handleWinChange(uploadFile) {
  switch (uploadFile.status) {
    case "success":
      winProc.value = "";
      break;
    case "ready":
      winFileName.value = uploadFile.name;
      winProc.value = "";
      winProcPercent.value = 0;
      break;
    case "uploading":
      winProc.value = "uploading";
      break;
    case "fail":
      winProc.value = "fail";
      break;
  }
}

function handleWinExceed(files) {
  winUploader.value?.clearFiles();
  winFileName.value = files[0].name;
  let file = files[0];
  file.uid = genFileId();
  winUploader.value?.handleStart(file);
}

function handleWinProgress(event, file, fileList) {
  switch (file.status) {
    case "success":
      winProc.value = "";
      break;
    case "ready":
      winProc.value = "";
      winProcPercent.value = 0;
      break;
    case "uploading":
      winProc.value = "uploading";
      winProcPercent.value = file.percentage;
      break;
    case "fail":
      winProc.value = "fail";
      break;
  }
}

function handleWinSuccess(response) {
  if (response.status == "success") {
    ClientVersion.value = response.data.client_version;
    wantWinVersion.value["version"] = response.data.client_version.win.version;
    wantWinVersion.value["url"] = response.data.client_version.win.url;
    winUploader.value?.clearFiles();
    winFileName.value = "";

    toastMsg.value = "Windows客户端发布成功";
    toastShow.value = true;
  } else {
    toastMsg.value = response.status.substring(6);
    toastShow.value = true;
  }
}

function handleWinError(err) {
  toastMsg.value = "上传文件失败：" + err;
  toastShow.value = true;
}

onMounted(() => {
  axios
    .get("/cockpit/api/setting/general")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        winUploadURL.value =
          "https://" + response.data["data"]["server_url"] + "/cockpit/api/publish/win";
        ClientVersion.value = response.data["data"]["client_version"];
        wantWinVersion.value = {};
        wantWinVersion.value["version"] =
          response.data["data"]["client_version"]["win"]["version"];
        wantWinVersion.value["url"] =
          response.data["data"]["client_version"]["win"]["url"];
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

function publishWin() {
  if (winExtURL.value) {
    axios
      .post("/cockpit/api/publish/win", {
        version: wantWinVersion.value["version"],
        url: wantWinVersion.value["url"],
      })
      .then(function (response) {
        // 处理成功情况
        if (response.data["status"] == "success") {
          toastMsg.value = "Windows客户端发布成功";
          toastShow.value = true;
          ClientVersion.value = response.data["data"]["client_version"];
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
    return;
  }
  winUploader.value?.submit();
}
</script>

<template>
  <div class="flex-1">
    <div
      class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center"
    >
      <h1 class="mr-2" tabindex="-1">客户端发布</h1>
    </div>
    <div class="text-gray-600 mt-3">
      <p>此页为发布蜃境客户端使用</p>
    </div>
    <div class="mt-6 space-y-6">
      <!---->
      <div>
        <header class="max-w-sm flex mt-4">
          <svg
            t="1679382174674"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="2682"
            width="28"
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
            Windows 客户端发布
          </h3>
        </header>
        <div
          class="rounded-md border border-stone-200 mt-4 mb-3 gap-2 max-w-sm bg-stone-50 p-2"
        >
          <div class="flex flex-col justify-start">
            <div class="w-full text-left font-bold max-w-xl text-gray-500">当前版本</div>
            <div class="w-full text-left max-w-xl text-gray-500">
              {{
                ClientVersion.win.version && ClientVersion.win.version != ""
                  ? ClientVersion.win.version
                  : "未设置"
              }}
            </div>
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              当前下载地址
            </div>
            <div class="w-full text-left max-w-xl break-all text-gray-500">
              {{
                ClientVersion.win.url && ClientVersion.win.url != ""
                  ? ClientVersion.win.url
                  : "未设置"
              }}
            </div>
          </div>
        </div>
        <div class="flex w-full max-w-sm">
          <p class="text-gray-600 max-w-sm min-w-fit">版本号</p>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                wantWinVersion.version == '' ||
                (winExtURL &&
                  (!isValidURL(wantWinVersion.url) ||
                    (wantWinVersion.version == ClientVersion.win.version &&
                      wantWinVersion.url == ClientVersion.win.url))) ||
                (!winExtURL && winFileName == '')
              "
              @click="publishWin"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              发布
            </button>
          </div>
        </div>
        <div
          :class="{
            'border-red-500 hover:border-red-700': wantWinVersion.version == '',
            'border-stone-200 hover:border-stone-400': wantWinVersion.version != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="wantWinVersion.version"
          />
        </div>

        <div class="flex flex-row w-full max-w-sm justify-start mt-2 space-x-2">
          <p class="text-gray-600 min-w-fit">上传文件</p>
          <input type="checkbox" class="toggle" v-model="winExtURL" />
          <p class="text-gray-600 min-w-fit">设置URL</p>
        </div>
        <label
          :class="{
            'swap-active': winExtURL,
          }"
          class="swap swap-flip block w-full max-w-sm text-md mt-2"
        >
          <div
            class="swap-on max-w-sm flex w-full border rounded-md relative overflow-hidden min-w-0"
            :class="{
              'border-red-500 hover:border-red-700': !isValidURL(wantWinVersion.url),
              'border-stone-200 hover:border-stone-400': isValidURL(wantWinVersion.url),
            }"
          >
            <input
              class="outline-none py-2 px-3 w-full h-9 font-mono text-sm text-ellipsis"
              v-model="wantWinVersion.url"
            />
          </div>
          <div class="swap-off flex w-full -mt-9">
            <el-upload
              ref="winUploader"
              drag
              :action="winUploadURL"
              :auto-upload="false"
              :show-file-list="false"
              :limit="1"
              :data="wantWinVersion"
              with-credentials="true"
              class="w-full"
              :on-exceed="handleWinExceed"
              :on-change="handleWinChange"
              :on-success="handleWinSuccess"
              :on-error="handleWinError"
              :on-progress="handleWinProgress"
            >
              <div class="el-upload__text">
                {{
                  winFileName && winFileName != ""
                    ? winFileName
                    : "将文件拖到此处，或点击上传"
                }}
              </div>
              <progress
                v-if="winProc != ''"
                :class="{
                  'progress-success': winProc == 'uploading',
                  'progress-error': winProc == 'fail',
                }"
                class="progress w-full"
                :value="winProcPercent"
                max="100"
              ></progress>
            </el-upload>
          </div>
        </label>
      </div>
      <!---->

      <!---->
    </div>
  </div>

  <!-- 菜单弹出提示框显示 -->
  <Teleport to="body"> </Teleport>
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
