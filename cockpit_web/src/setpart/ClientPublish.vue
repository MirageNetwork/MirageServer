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
const uploadURL = ref("");

const wantNaviVersion = ref("");
const naviAmd64Uploader = ref(null);
const naviAarch64Uploader = ref(null);
const naviProc = ref("");
const naviProcPercent = ref(0);

const wantWinVersion = ref({});
const winExtURL = ref(false);
const winUploader = ref(null);
const winFileName = ref("");
const winProc = ref("");
const winProcPercent = ref(0);

function handleNaviChange(uploadFile) {
  switch (uploadFile.status) {
    case "success":
      naviProc.value = "";
      break;
    case "ready":
      naviProc.value = "";
      naviProcPercent.value = 0;
      break;
    case "uploading":
      naviProc.value = "uploading";
      break;
    case "fail":
      naviProc.value = "fail";
      break;
  }
}

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

function handleNaviAmd64Exceed(files) {
  naviAmd64Uploader.value?.clearFiles();
  let file = files[0];
  file.uid = genFileId();
  naviAmd64Uploader.value?.handleStart(file);
}
function handleNaviAarch64Exceed(files) {
  naviAarch64Uploader.value?.clearFiles();
  let file = files[0];
  file.uid = genFileId();
  naviAarch64Uploader.value?.handleStart(file);
}

function handleWinExceed(files) {
  winUploader.value?.clearFiles();
  winFileName.value = files[0].name;
  let file = files[0];
  file.uid = genFileId();
  winUploader.value?.handleStart(file);
}

function handleNaviProgress(event, file, fileList) {
  switch (file.status) {
    case "success":
      naviProc.value = "";
      break;
    case "ready":
      naviProc.value = "";
      naviProcPercent.value = 0;
      break;
    case "uploading":
      naviProc.value = "uploading";
      naviProcPercent.value = file.percentage;
      break;
    case "fail":
      naviProc.value = "fail";
      break;
  }
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

function handleNaviSuccess(response) {
  if (response.status == "success") {
    ClientVersion.value = response.data.client_version;
    wantNaviVersion.value = "";
    naviAmd64Uploader.value?.clearFiles();
    naviAarch64Uploader.value?.clearFiles();

    toastMsg.value = "司南客户端发布成功";
    toastShow.value = true;
  } else {
    toastMsg.value = response.status.substring(6);
    toastShow.value = true;
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

function handleNaviError(err) {
  toastMsg.value = "上传文件失败：" + err;
  toastShow.value = true;
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
        uploadURL.value =
          "https://" + response.data["data"]["server_url"] + "/cockpit/api/publish/";
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
            t="1680440667504"
            class="icon"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="5686"
            width="28"
            height="28"
          >
            <path
              d="M347.8 745.1l-164 54.7V349.1l164-54.7zM839.8 809.7l-164.1 54.7V413.7l164.1-54.7z"
              fill="#F3F5CD"
              p-id="5687"
            ></path>
            <path
              d="M416.3 432.7c36.1 1.6 84.5 32.5 117.8 75.1 35.5 45.4 49.5 97.8 39.5 147.5-5.3 26.4-15.6 43.5-31.3 52.1-25.4 14-57.6 1.8-86.1-9-21-7.9-42.6-16.2-49.9-8.3-1.4 1.5-1.6 2.2-0.5 4.3 6.2 11.5 36.8 27.6 77.9 40.9 38.3 12.4 76.9 19.6 91.7 17.1 53.2-8.9 93.9-49.1 94.3-49.5 1.7-1.7 3.9-2.7 6.5-2.7V413.7L347.8 294.4v190.1c4.4-12.4 10.6-24.1 19.4-33.2 12.9-13.3 29.4-19.5 49.1-18.6z"
              fill="#C6C49E"
              p-id="5688"
            ></path>
            <path
              d="M568.5 772c-23.1 0-60.5-8.7-90.7-18.5-28-9.1-76.7-27.7-88.7-50-4.9-9.1-3.7-18.7 3.3-26.2 16-17.2 42.5-7.1 70.5 3.5 24.6 9.3 52.5 19.9 70.2 10.1 10.4-5.8 17.8-19 21.9-39.3 8.9-44.2-3.8-91.1-35.8-132.1-33.6-43-77.6-66.7-103.7-67.9-14.2-0.7-25.6 3.5-34.6 12.8-28.3 29.2-23.6 96.2-23.6 96.9v0.7c0 5.1-4 9.3-9.1 9.5h-0.4V745.1l328.4 119.3V722.8c-14.7 12.7-51.4 40.7-97.7 48.4-2.9 0.5-6.2 0.8-10 0.8z"
              fill="#C6C49E"
              p-id="5689"
            ></path>
            <path
              d="M339 527.5c-11.4-15.8-27.9-35.8-48.7-58.9-11.1-12.3-24-24.2-36-18.6-15.7 7.4-12.4 35-4.2 91.7 5.1 35.1 10.9 74.8 6.4 95-1.1 5-3.3 11.2-7.4 12.2-12.8 2.9-42.2-23.9-58.1-43.2l-14.6 12.1c7.8 9.5 48.3 56.4 77.1 49.5 7.5-1.8 17.3-7.8 21.5-26.4 5.3-23.5-0.5-63.4-6.2-101.8-3.5-23.9-8.6-59-6.1-70.2 2.3 1.5 6.6 4.9 13.5 12.5 55.5 61.9 61.7 80.1 62.2 82 0.6 4.7 4.6 8.3 9.4 8.3v-87c-5.1 14.2-7.6 29.5-8.8 42.8z"
              fill="#F4CD7D"
              p-id="5690"
            ></path>
            <path
              d="M669.7 703.1c-0.4 0.4-41.1 40.6-94.3 49.5-14.9 2.5-53.5-4.7-91.7-17.1-41.1-13.4-71.7-29.4-77.9-40.9-1.1-2.1-0.9-2.8 0.5-4.3 7.3-7.9 29 0.4 49.9 8.3 28.5 10.8 60.7 23 86.1 9 15.7-8.7 26-25.8 31.3-52.1 10-49.7-4-102.1-39.5-147.5-33.3-42.6-81.7-73.5-117.8-75.1-19.6-0.9-36.2 5.3-49 18.6-8.9 9.1-15.1 20.8-19.4 33.2v87h0.4c5.1-0.2 9.1-4.4 9.1-9.5v-0.7c-0.1-0.7-4.7-67.7 23.6-96.9 9-9.3 20.3-13.5 34.6-12.8 26.1 1.2 70.2 24.9 103.7 67.9 32 41 44.8 87.9 35.8 132.1-4.1 20.3-11.5 33.6-21.9 39.3-17.7 9.8-45.6-0.8-70.2-10.1-28.1-10.6-54.6-20.7-70.5-3.5-7 7.5-8.2 17.1-3.3 26.2 12 22.3 60.8 40.9 88.7 50 30.2 9.8 67.6 18.5 90.7 18.5 3.7 0 7.1-0.2 10-0.7 46.2-7.7 83-35.7 97.7-48.4v-22.8c-2.7 0.1-4.9 1.1-6.6 2.8z"
              fill="#CCA766"
              p-id="5691"
            ></path>
            <path
              d="M846.5 499l-12.4-14.3c-3.1 2.7-76.6 67.2-85.2 138.3-4.1 33.7 3 55.4 8.1 71.2 5.6 17.1 7.4 22.9-4.9 33.5-24.3 21-68.3-24.1-68.7-24.5-1.8-1.9-4.2-2.9-6.8-3h-0.3v22.5c16 14.5 57.3 46 88.2 19.4 21.9-18.9 16.4-35.8 10.6-53.8-4.9-14.9-10.9-33.5-7.3-63.1 7.5-63.8 77.9-125.6 78.7-126.2z"
              fill="#F4CD7D"
              p-id="5692"
            ></path>
            <path
              d="M676.1 498.8c-1.9-1.3-4-2.6-6.3-2.4s-4.4 2.6-3.5 4.7h-1.8c-7.9-7.8-18.9-12.4-30.1-12.4-2.5 0-5.2 0.2-7.3 1.6-3.5 2.3-4.6 6.9-5.3 11.1-1.3 7.3-2.5 15.1 0.2 22 4.7 12 19.1 17.1 26.3 27.7 8.4 12.2 5.6 28.7 1.3 42.8-2.2 7.3-4.7 15-2.8 22.4 2.3 8.8 10.6 15.1 19.4 17.5 3.3 0.9 6.6 1.4 10 1.5l-0.1-136.5c0.1 0 0 0 0 0z"
              fill="#68BBD8"
              p-id="5693"
            ></path>
            <path
              d="M693.4 634.7c2.9-0.2 6-0.6 8.2-2.4 3.2-2.6 3.5-7.3 4.3-11.4 1.5-8.7 5.9-16.9 12.3-23 2.2-2.1 4.8-4 7.8-4.4 4.8-0.7 9.2 2.6 13.8 4.2 11.5 3.9 24.5-3.4 30.7-13.8s7.2-23.1 7.4-35.2c0.1-8.3-0.1-16.8-3-24.6-2.9-7.8-9.1-14.8-17.2-16.7-14.2-3.3-26.5 9.3-36.4 19.9-2.4 2.5-5.1 5.2-8.5 5.5-5.5 0.4-9.3-5-12.4-9.5-6.5-9.5-14.7-17.8-24.1-24.3v136.3c5.6 0.4 11.4-0.1 17.1-0.6z"
              fill="#8DDCFF"
              p-id="5694"
            ></path>
            <path
              d="M340.1 291.5c-22.4-2.7-52.8 5.9-72.9 18-11.6 3-21.8 7.9-24.8 16.9-11.5 34.3 32.4 65.5 62 50.3 7 0.3 10.4 8.7 18 13.1 7.8 4.6 16.7 6.1 25.5 5.2V292.7c-2.9-0.6-5.5-1-7.8-1.2zM340 591c-2.2 1.6-4.6 3.5-5.3 6.1-0.7 2.7 1.2 6.1 3.9 5.8l-0.7 2.1c-12 5.8-21.7 16.4-26.2 28.9-1 2.9-1.8 6-1.1 8.9 1.2 4.9 6 8 10.4 10.5 7.8 4.4 16 8.9 24.9 8.7 0.7 0 1.3-0.1 1.9-0.2v-75.8c-2.7 1.5-5.3 3.2-7.8 5z"
              fill="#A9D335"
              p-id="5695"
            ></path>
            <path
              d="M397.1 319.8c-5.7-15.3-31.7-23.8-49.3-27.2v102.3c11-1.1 21.8-6.2 29.8-13.8 13.4-12.8 26.5-42.6 19.5-61.3zM436.6 659.5c7.3 5.4 15 11.4 24.1 12.3 10.9 1 21.3-5.8 27.6-14.8 6.3-8.9 9.3-19.8 12.1-30.3 0.9-3.4 1.8-7 0.6-10.3-1.7-4.7-6.8-7-11.2-9.4-9.2-5.2-16.7-13.5-21-23.2-1.5-3.3-2.6-7-1.8-10.6 1.2-5.7 6.7-9.3 10.3-13.9 9.1-11.4 6.1-29-3.1-40.3-9.2-11.3-23.1-17.5-36.8-22.7-9.3-3.5-19-6.7-29-6.6-10 0.2-20.4 4.3-25.8 12.6-9.5 14.7-0.3 33.7 7.7 49.2 1.9 3.7 3.8 7.8 2.7 11.9-1.7 6.3-9.4 8.5-15.8 10.1-10.4 2.7-20.4 6.9-29.6 12.3v75.8c14.5-1.4 25.7-14.6 40-18.3 17.4-4.3 34.8 5.6 49 16.2z"
              fill="#7D992E"
              p-id="5696"
            ></path>
            <path
              d="M676.2 877c-1.5 0-2.9-0.3-4.3-0.8L347.6 758.4l-159.9 53.3c-3.8 1.3-8 0.6-11.3-1.7-3.3-2.4-5.2-6.2-5.2-10.2V349.1c0-5.4 3.5-10.2 8.6-11.9l164-54.7c2.7-0.9 5.6-0.9 8.3 0.1l324.3 117.8 159.9-53.3c3.8-1.3 8-0.6 11.3 1.7 3.3 2.4 5.2 6.2 5.2 10.2v450.7c0 5.4-3.5 10.2-8.6 11.9l-164 54.7c-1.3 0.5-2.7 0.7-4 0.7zM347.8 732.5c1.5 0 2.9 0.3 4.3 0.8L676.4 851.1l151.3-50.5V376.4l-147.5 49.2c-2.7 0.9-5.6 0.9-8.3-0.1L347.6 307.7l-151.3 50.5v424.2l147.5-49.2c1.3-0.5 2.7-0.7 4-0.7z"
              fill="#592900"
              p-id="5697"
            ></path>
            <path
              d="M486 562a60.1 17.4 0 1 0 120.2 0 60.1 17.4 0 1 0-120.2 0Z"
              fill="#999881"
              p-id="5698"
            ></path>
            <path
              d="M546.1 157c-72 0-130.3 58.3-130.3 130.3 0 104.2 130.3 257.4 130.3 257.4s130.3-157.5 130.3-257.4c0-72-58.3-130.3-130.3-130.3z m0 170.2c-24.4 0-44.3-19.8-44.3-44.3 0-24.4 19.8-44.3 44.3-44.3s44.3 19.8 44.3 44.3c0 24.5-19.8 44.3-44.3 44.3z"
              fill="#F04E2F"
              p-id="5699"
            ></path>
            <path
              d="M546.1 554.7c-2.9 0-5.7-1.3-7.7-3.5-5.4-6.4-132.6-157.2-132.6-263.9 0-77.4 63-140.3 140.3-140.3s140.3 63 140.3 140.3c0 102.3-127.2 257.2-132.6 263.8-1.8 2.2-4.6 3.6-7.7 3.6 0.1 0 0 0 0 0z m0-387.7c-66.3 0-120.2 53.9-120.2 120.2 0 84.9 92.4 206.7 120.1 241.4 27.6-35.3 120.3-160 120.3-241.4 0.1-66.2-53.9-120.2-120.2-120.2z m0 170.3c-30 0-54.3-24.4-54.3-54.3 0-30 24.4-54.3 54.3-54.3 30 0 54.3 24.4 54.3 54.3s-24.3 54.3-54.3 54.3z m0-88.6c-18.9 0-34.2 15.3-34.2 34.2s15.3 34.2 34.2 34.2 34.2-15.4 34.2-34.2-15.3-34.2-34.2-34.2z"
              fill="#592900"
              p-id="5700"
            ></path>
            <path
              d="M456.9 267c-0.7 0-1.5-0.1-2.2-0.3-4-1.2-6.2-5.4-5-9.4 7.4-24.4 23.4-45.4 45.2-58.9 3.5-2.2 8.2-1.1 10.4 2.4s1.1 8.2-2.4 10.4a89.963 89.963 0 0 0-38.7 50.5c-1.1 3.2-4.1 5.3-7.3 5.3z"
              fill="#FFFFFF"
              p-id="5701"
            ></path>
            <path
              d="M452.9 301.8c-3.9 0-7.3-3.1-7.5-7-0.2-2.3-0.2-4.6-0.2-7 0-3.9 0.2-7.9 0.7-11.9 0.5-4.1 4.2-7.1 8.3-6.7 4.1 0.5 7.1 4.2 6.7 8.3-0.4 3.4-0.6 6.8-0.6 10.2 0 2 0.1 4 0.2 6 0.3 4.2-2.9 7.7-7 8-0.2 0.1-0.4 0.1-0.6 0.1z"
              fill="#FFFFFF"
              p-id="5702"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">司南发布</h3>
        </header>
        <div
          class="rounded-md border border-stone-200 mt-4 mb-3 gap-2 max-w-sm bg-stone-50 p-2"
        >
          <div class="flex flex-col justify-start">
            <div class="w-full text-left font-bold max-w-xl text-gray-500">当前版本</div>
            <div class="w-full text-left max-w-xl text-gray-500">
              x86_64:
              {{
                ClientVersion.naviAmd64 && ClientVersion.naviAmd64 != ""
                  ? ClientVersion.naviAmd64
                  : "未设置"
              }}
              <br />
              aarch64:
              {{
                ClientVersion.naviAarch64 && ClientVersion.naviAarch64 != ""
                  ? ClientVersion.naviAarch64
                  : "未设置"
              }}
            </div>
            <progress
              v-if="naviProc != ''"
              :class="{
                'progress-success': naviProc == 'uploading',
                'progress-error': naviProc == 'fail',
              }"
              class="progress w-full"
              :value="naviProcPercent"
              max="100"
            ></progress>
          </div>
        </div>
        <div class="flex w-full max-w-sm space-x-2 justify-around items-center mb-3">
          <p class="text-gray-600 max-w-sm min-w-fit">版本号</p>
          <div
            :class="{
              'border-red-500 hover:border-red-700': wantNaviVersion == '',
              'border-stone-200 hover:border-stone-400': wantNaviVersion != '',
            }"
            class="max-w-sm flex w-full border rounded-md relative overflow-hidden min-w-0"
          >
            <input
              class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
              v-model="wantNaviVersion"
            />
          </div>
        </div>
        <div class="w-full max-w-sm flex justify-around">
          <el-upload
            :disabled="(!naviProc && naviProc != '') || wantNaviVersion == ''"
            ref="naviAmd64Uploader"
            :action="uploadURL + 'navi_x86_64'"
            :show-file-list="false"
            :limit="1"
            :data="{ version: wantNaviVersion }"
            :with-credentials="true"
            class="w-1/2"
            :on-exceed="handleNaviAmd64Exceed"
            :on-change="handleNaviChange"
            :on-success="handleNaviSuccess"
            :on-error="handleNaviError"
            :on-progress="handleNaviProgress"
          >
            <button
              :disabled="(!naviProc && naviProc != '') || wantNaviVersion == ''"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit w-full"
            >
              上传x86_64版本
            </button>
          </el-upload>
          <el-upload
            :disabled="(!naviProc && naviProc != '') || wantNaviVersion == ''"
            ref="naviAarch64Uploader"
            :action="uploadURL + 'navi_aarch64'"
            :show-file-list="false"
            :limit="1"
            :data="{ version: wantNaviVersion }"
            :with-credentials="true"
            class="w-1/2"
            :on-exceed="handleNaviAarch64Exceed"
            :on-change="handleNaviChange"
            :on-success="handleNaviSuccess"
            :on-error="handleNaviError"
            :on-progress="handleNaviProgress"
          >
            <button
              :disabled="(!naviProc && naviProc != '') || wantNaviVersion == ''"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit w-full"
            >
              上传aarch64版本
            </button>
          </el-upload>
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
                (!winProc && winProc != '') ||
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
              :action="uploadURL + 'win'"
              :auto-upload="false"
              :show-file-list="false"
              :limit="1"
              :data="wantWinVersion"
              :with-credentials="true"
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
