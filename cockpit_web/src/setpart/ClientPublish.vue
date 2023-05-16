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

const wantLinuxRepo = ref("");
const wantLinuxRepoCred = ref("");

const wantWinVersion = ref({});
const winExtURL = ref(false);
const winUploader = ref(null);
const winFileName = ref("");
const winProc = ref("");
const winProcPercent = ref(0);

const wantIOSStoreVersion = ref({});
const wantIOSTestVersion = ref({});

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

var getLinuxBuildStateIntID;

function getLinuxBuildState() {
  axios
    .get("/cockpit/api/publish")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        uploadURL.value = response.data["data"]["upload_url"];
        ClientVersion.value = response.data["data"]["client_version"];
        if (ClientVersion.value["linux"]["buildst"] != "正在进行") {
          clearInterval(getLinuxBuildStateIntID);
        }
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
}

onMounted(() => {
  axios
    .get("/cockpit/api/publish")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        uploadURL.value = response.data["data"]["upload_url"];
        ClientVersion.value = response.data["data"]["client_version"];

        wantLinuxRepo.value = response.data["data"]["client_version"]["linux"]["url"];

        wantWinVersion.value = {};
        wantWinVersion.value["version"] =
          response.data["data"]["client_version"]["win"]["version"];
        wantWinVersion.value["url"] =
          response.data["data"]["client_version"]["win"]["url"];
        wantIOSStoreVersion.value = {};
        wantIOSStoreVersion.value["version"] =
          response.data["data"]["client_version"]["ios_store"]["version"];
        wantIOSStoreVersion.value["url"] =
          response.data["data"]["client_version"]["ios_store"]["url"];
        wantIOSTestVersion.value = {};
        wantIOSTestVersion.value["version"] =
          response.data["data"]["client_version"]["ios_test"]["version"];
        wantIOSTestVersion.value["url"] =
          response.data["data"]["client_version"]["ios_test"]["url"];
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

function publishLinux() {
  axios
    .post("/cockpit/api/publish/linux", {
      version: wantLinuxRepoCred.value,
      url: wantLinuxRepo.value,
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        toastMsg.value = "Linux已提交构建";
        toastShow.value = true;
        getLinuxBuildStateIntID = setInterval(() => {
          getLinuxBuildState();
        }, 3000);
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

function publishIOSToStore() {
  axios
    .post("/cockpit/api/publish/ios_store", {
      version: wantIOSStoreVersion.value["version"],
      url: wantIOSStoreVersion.value["url"],
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        toastMsg.value = "iOS客户端AppStore发布成功";
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
function publishIOSToTestflight() {
  axios
    .post("/cockpit/api/publish/ios_test", {
      version: wantIOSTestVersion.value["version"],
      url: wantIOSTestVersion.value["url"],
    })
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        toastMsg.value = "iOS客户端TestFlight发布成功";
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
            :action="uploadURL + '/navi_x86_64'"
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
            :action="uploadURL + '/navi_aarch64'"
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
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            width="28"
            height="28"
          >
            <path
              d="M525.2 198.3c-8.6 5.6-15.2 13.8-18.9 23.4-3.8 12.4-3.2 25.6 1.5 37.7 3.9 12.7 11.7 23.8 22.2 31.8 5.4 3.8 11.6 6.2 18.2 7 6.6 0.8 13.2-0.3 19.1-3.3 7-3.9 12.6-10 15.9-17.3 3.2-7.4 5-15.3 5.2-23.3 0.7-10.2-0.6-20.4-3.8-30.1-3.5-10.6-10.3-19.7-19.5-25.9-4.7-3-9.9-5-15.4-5.8-5.5-0.8-11.1-0.2-16.3 1.8-2.9 1.2-5.7 2.7-8.3 4.5"
              fill="#FFFFFF"
            ></path>
            <path
              d="M810.2 606.5c-5.1-28.3-13.1-56-23.8-82.6-7.3-19.8-17.2-38.6-29.5-55.8-12.4-16.5-28.1-30.4-40.2-47.1-6.4-8.7-11.8-18.4-18.5-26.9-2.7-5.6-5.3-11.2-7.9-16.8-8-17.5-15.3-35.4-24.8-52-1.5-2.6-3.1-5.2-4.6-7.7-1.2-16-2.9-32-3.8-48 0.7-32.1-2-64.3-8.1-95.9-4.2-15.1-10.6-29.6-19-42.8-9.8-15.6-22.4-29.2-37.2-40.1-24.1-17.1-52.9-26.3-82.4-26.4-21.7-0.5-43.2 4.4-62.5 14.4-20.3 11.1-36.7 28.2-47 48.9-9.6 20.9-14.7 43.5-15 66.5-0.8 22.6 1.3 45 2.2 67.6 0.9 23.4 0.4 46.9 2.3 70.3 0.6 7.5 1.5 15 1.5 22.6 0 3.8-0.2 7.6-0.3 11.3l-0.3 0.8c-10.2 17.3-21.5 34-33.8 49.9-8.6 10.9-17.2 21.7-25.9 32.4-11.3 12.7-20.9 26.8-28.5 42-5.1 13.2-9.2 26.8-12.4 40.6l-0.3 1.1c-4.8 15.9-10.8 31.3-18 46.2-0.7 1.4-1.4 2.9-2 4.2-4.3 8.9-8.8 17.8-13.5 26.5l-5.4 10.1c-3.4 6.1-6.4 12.4-9 18.8-1.5 3.9-2.7 7.9-3.4 12-1.3 8.7-0.7 17.5 1.6 25.9 0.5 2.1 1.2 4.2 1.9 6.3 2.2 6.2 4.8 12.3 7.9 18.1 1.4 2.7 2.9 5.3 4.3 8l1.3 1.9c1.4 2.5 2.9 5 4.4 7.4l0.2 0.3c1.7 2.8 3.6 5.5 5.4 8.2l0.3 0.4c1.9 2.6 3.8 5.3 5.8 7.9 7.4 28.9 21 55.8 39.7 79-2.9 5.1-5.5 10.1-8.4 15.1-10.2 14.8-18.6 30.7-25.1 47.4-2.7 8.6-3.4 17.7-1.9 26.6 1.4 9 6 17.1 13 23 4.7 3.6 10.1 6.1 15.8 7.3 5.7 1.2 11.6 1.8 17.5 1.5 22.2-1.7 44.2-6.1 65.4-12.9 12.8-3.4 25.6-6.4 38.6-9 13.5-3.1 27.2-5 41-5.6 3.4 0.1 6.8-0.1 10.1-0.3 9.4 1 18.8 1.4 28.3 1l3.5-0.2c2.4 0.3 4.9 0.4 7.4 0.6 16.6 0.9 33.1 2.6 49.5 5.1 14.4 2.2 28.8 5 43 8.5 21.9 6.6 44.4 11 67.3 12.9 6 0.3 12-0.2 18-1.4 5.9-1.2 11.5-3.8 16.3-7.4 7-5.8 11.6-13.9 13.1-22.9 1.5-8.9 0.8-18-1.9-26.6-6.6-16.7-15.1-32.6-25.5-47.3-3.6-6.1-7-12.4-10.6-18.5 15.5-17.3 29.2-36.3 40.7-56.5 7 0.4 13.9-0.4 20.6-2.6 17.5-5.9 32.7-17.3 43.3-32.5 3.2-4.5 5.7-9.5 7.2-14.8 6.9-10.7 11.6-22.7 13.8-35.3 3.2-20.8 2.7-42.1-1.5-62.7h-0.2z m0 0"
              fill="#020204"
            ></path>
            <path
              d="M425.6 323.2c-3.1 4-5.3 8.7-6.4 13.6-1.1 4.9-1.8 10-1.9 15 0.3 10.1-0.5 20.2-2.5 30.1-3.5 10.3-8.8 19.8-15.6 28.3-11.7 14.7-20.9 31.2-27.2 48.8-3.2 10.9-4.3 22.3-3.1 33.7-12.1 17.9-22.6 36.9-31.3 56.7-13.4 29.9-22 61.8-25.5 94.4-4.3 40.1 1.6 80.6 17 117.8 11.3 26.8 28.5 50.8 50.3 70.1 11.2 9.7 23.5 17.9 36.7 24.4 46.7 22.8 101.4 22.3 147.6-1.4 23.1-13.5 44.2-30.2 62.6-49.5 11.9-10.8 22.5-22.9 31.8-36.1 15.5-26.9 24.6-57.1 26.5-88.1 9.6-53.6 3.7-108.8-16.9-159.2-8.1-16.8-18.8-32.2-31.8-45.6a252.5 252.5 0 0 0-20.2-68c-7.2-15.5-15.9-30.3-22.6-46.2-2.7-6.5-5.1-13.1-8.1-19.4-2.9-6.4-6.9-12.3-11.8-17.3-5.3-4.9-11.6-8.6-18.5-10.7-6.9-2.2-14-3.4-21.2-3.6-14.4-0.7-28.9 1.1-43.1 0.6-11.5-0.5-22.8-2.5-34.3-1.8-5.7 0.3-11.4 1.4-16.7 3.5-5.4 2.1-10.1 5.5-13.8 10m4.6-125.1c-5.4 0.4-10.5 2.7-14.4 6.4-3.9 3.7-6.8 8.4-8.4 13.5-2.7 10.4-3.4 21.3-1.9 32 0.2 9.7 1.9 19.4 5.1 28.6 1.8 4.5 4.4 8.7 7.8 12.2 3.4 3.5 7.7 6.1 12.4 7.3 4.5 1.1 9.2 0.9 13.5-0.5 4.3-1.4 8.3-3.8 11.5-7 4.7-4.8 8.1-10.7 9.8-17.1 1.7-6.4 2.5-13.1 2.3-19.8 0-8.3-1.3-16.6-3.8-24.6s-6.8-15.3-12.6-21.4c-2.8-2.9-6-5.4-9.6-7.2-3.7-1.7-7.7-2.6-11.7-2.4m95 0c-8.6 5.6-15.2 13.8-18.9 23.4-3.8 12.4-3.2 25.6 1.5 37.7 3.9 12.7 11.7 23.8 22.2 31.8 5.4 3.8 11.6 6.2 18.2 7 6.6 0.8 13.2-0.3 19.1-3.3 7-3.9 12.6-10 15.9-17.3 3.2-7.4 5-15.3 5.2-23.3 0.7-10.2-0.6-20.4-3.8-30.1-3.5-10.6-10.3-19.7-19.5-25.9-4.7-3-9.9-5-15.4-5.8-5.5-0.8-11.1-0.2-16.3 1.8-2.9 1.2-5.7 2.7-8.3 4.5"
              fill="#FFFFFF"
            ></path>
            <path
              d="M544.5 223.6c-3.2 0.2-6.2 1.2-8.9 2.9s-5 4-6.8 6.6c-3.4 5.3-5.3 11.5-5.4 17.9-0.3 4.7 0.4 9.5 1.9 14s4.3 8.5 7.9 11.5c3.8 3.1 8.4 4.9 13.3 5.2 4.9 0.2 9.7-1.1 13.7-3.9 3.2-2.3 5.8-5.2 7.6-8.7 1.8-3.4 2.9-7.2 3.4-11 1-6.8-0.2-13.8-3.2-19.9-3.1-6.2-8.4-10.9-14.8-13.4-2.8-1.1-5.7-1.5-8.7-1.4"
              fill="#020204"
            ></path>
            <path
              d="M430.2 198.3c-5.4 0.4-10.5 2.7-14.4 6.4-3.9 3.7-6.8 8.4-8.4 13.5-2.7 10.4-3.4 21.3-1.9 32 0.2 9.7 1.9 19.4 5.1 28.6 1.8 4.6 4.4 8.7 7.8 12.2 3.4 3.5 7.7 6.1 12.4 7.3 4.5 1.1 9.2 0.9 13.5-0.5 4.3-1.4 8.3-3.8 11.5-7 4.7-4.8 8.1-10.7 9.8-17.1 1.7-6.4 2.5-13.1 2.3-19.8 0-8.3-1.3-16.6-3.8-24.6s-6.8-15.3-12.6-21.4c-2.8-2.9-6-5.4-9.6-7.2-3.7-1.7-7.7-2.6-11.7-2.4"
              fill="#FFFFFF"
            ></path>
            <path
              d="M417.3 242.8c-1.3 6.7-1 13.7 1.1 20.2 1.6 4.3 4 8.2 7.2 11.5 2 2.2 4.3 4.1 7 5.4 2.7 1.4 5.7 1.8 8.7 1.1 2.7-0.7 5-2.3 6.7-4.5 1.7-2.2 2.9-4.7 3.7-7.3 2.3-7.8 2.1-16.1-0.4-23.9-1.6-5.7-4.7-10.9-9.1-14.8-2.1-1.8-4.7-3.2-7.4-3.9-2.8-0.7-5.7-0.5-8.4 0.7-2.8 1.4-5.1 3.7-6.5 6.5-1.4 2.8-2.3 5.8-2.7 8.9"
              fill="#020204"
            ></path>
            <path
              d="M404.6 326.9c0.2 0.9 0.5 1.8 1 2.5 0.9 1.4 2 2.5 3.4 3.4 1.3 0.9 2.6 1.7 3.9 2.5 6.9 4.7 13 10.5 17.9 17.3 6 9.4 13.5 17.8 22 25 6.5 4.5 14.1 7.2 22 7.9 9.2 0.7 18.5-0.4 27.4-3.2 8.2-2.4 16.1-5.8 23.5-10.3 12.7-10.2 26.3-19.2 40.7-26.7 3.4-1.2 6.8-2.1 10-3.6 3.3-1.4 6.1-3.8 7.8-7 1.1-3.2 1.8-6.6 1.9-10 0.5-3.6 1.7-7.1 2.3-10.7 0.8-3.6 0.5-7.3-0.8-10.8-1.4-2.7-3.6-4.9-6.3-6.3-2.7-1.3-5.7-2.1-8.7-2.2-6.1 0.2-12.1 0.8-18 1.8-8 0.7-16-0.3-24 0-9.9 0.3-19.8 2.5-29.8 2.9-11.4 0.6-22.7-1.2-34.1-1.7-4.9-0.3-9.9-0.1-14.8 0.7-4.9 0.7-9.6 2.5-13.7 5.3-3.8 3-7.3 6.2-10.7 9.6-1.8 1.6-3.8 3-5.9 4.1-2.2 1.1-4.5 1.7-7 1.6-1.2-0.2-2.5-0.2-3.7 0-0.7 0.3-1.4 0.7-1.9 1.2l-1.5 1.8c-1 1.5-1.9 3.1-2.6 4.7"
              fill="#D99A03"
            ></path>
            <path
              d="M429.7 301.7c-4 2.4-7.9 5-11.8 7.7-2.1 1.3-3.8 3-5.1 5.1-0.7 1.6-1 3.3-0.9 5 0.1 1.7 0.1 3.4 0 5.1-0.1 1.1-0.5 2.3-0.5 3.5 0 0.6 0 1.2 0.2 1.7 0.2 0.6 0.4 1.1 0.8 1.5 0.5 0.5 1.2 0.9 2 1.1 0.7 0.2 1.5 0.3 2.3 0.5 3.5 1 6.7 2.9 9.3 5.4 2.7 2.4 5.1 5.2 8 7.5 8 6 17.7 9.1 27.6 9 9.9-0.2 19.7-1.6 29.2-4.1 7.5-1.6 14.9-3.6 22.1-6.1 11.2-4.2 21.5-10.3 30.4-18.2 3.9-3.8 8-7.2 12.4-10.3 4-2.5 8.7-4.2 12.7-6.6 0.4-0.2 0.7-0.5 1.1-0.7 0.3-0.3 0.6-0.6 0.8-1 0.3-0.7 0.3-1.5 0-2.2-0.2-0.7-0.5-1.3-0.9-1.8-0.5-0.6-1.1-1.2-1.7-1.7-4.6-3.4-10.1-5.3-15.8-5.5-5.8-0.4-11.3 0-16.9-1.1-5.2-1.1-10.3-2.6-15.3-4.4-5.3-1.7-10.7-3-16.3-3.9-13-2.1-26.2-1.8-39.1 1-12.1 2.7-23.8 7.3-34.6 13.5"
              fill="#604405"
            ></path>
            <path
              d="M428.4 288.1c-5.8 3.9-11 8.7-15.5 14.1-2.6 3-4.7 6.5-6.1 10.3-0.9 3-1.5 6.1-2 9.2-0.3 1.1-0.5 2.3-0.5 3.5 0 0.6 0.1 1.2 0.3 1.7 0.2 0.6 0.5 1.1 0.9 1.5 0.7 0.7 1.6 1.1 2.6 1.3 0.9 0.2 1.9 0.2 2.9 0.3 4.4 0.7 8.5 2.5 12.1 5.1 3.6 2.5 7 5.4 10.7 7.8 8.4 5 18 7.7 27.8 7.9 9.8 0.2 19.5-0.8 29-2.9 7.6-1.4 15.1-3.5 22.4-6.3 10.9-4.7 21.1-10.8 30.4-18.2 4.3-3.2 8.5-6.6 12.4-10.3 1.3-1.3 2.6-2.6 4-3.8 1.4-1.2 3-2.1 4.7-2.7 2.7-0.7 5.5-0.8 8.3-0.1 2 0.5 4.1 0.7 6.2 0.7 1.1 0 2.1-0.2 3.1-0.5 1-0.4 1.9-1 2.5-1.8 0.9-1.1 1.3-2.4 1.3-3.8s-0.4-2.7-1.1-3.9c-1.5-2.3-3.8-4.1-6.3-5.1-3.5-1.4-7.1-2.5-10.8-3.2-11.3-2.7-22.3-6.7-32.7-11.9-5.2-2.6-10.1-5.4-15.3-8.1-5.2-2.9-10.6-5.4-16.2-7.2-12.9-3.5-26.6-2.9-39.1 1.8-14 4.9-26.5 13.4-36.1 24.7"
              fill="#F5BD0C"
            ></path>
            <path
              d="M493.5 272.2c0.7 2.3 4.3 1.9 6.4 2.9 2.1 1 3.3 2.9 5.3 3.1 2.1 0.2 5-0.7 5.3-2.6 0.4-2.6-3.4-4.2-5.8-5.1-3.2-1.5-6.8-1.6-10-0.2-0.7 0.3-1.4 1.2-1.2 1.9z m-34.4-1.2c-2.7-0.9-7.1 3.8-5.8 6.3 0.4 0.7 1.6 1.5 2.4 1.1 0.8-0.4 2.3-3.1 3.6-4 1-0.8 0.8-3.1-0.2-3.4z m0 0"
              fill="#CD8907"
            ></path>
            <path
              d="M887.7 829.8c-2 5.2-4.9 10-8.5 14.3-8.4 9-18.6 16.2-29.8 21.2-19 8.8-37.5 18.6-55.5 29.3-11.7 7.8-22.6 16.6-32.7 26.4-8.3 8.7-17.2 16.7-26.6 24.2-9.8 7.2-21.1 12.1-33.1 14-14.7 1.9-29.6-0.4-43.1-6.5-9.7-3.7-18.1-10.2-24-18.8-5-9.2-7.3-19.5-6.8-29.9 0.6-18.3 2.8-36.5 6.6-54.5 2.6-15 5.2-30 6.8-45.1 2.8-27.6 3.1-55.3 1-82.9-0.5-4.6-0.5-9.3 0-13.9 0.6-9.4 8.5-16.6 18-16.5 4.3-0.1 8.6 0.3 12.8 1.1 10 1.2 20 2.9 29.8 5.2 6.1 1.6 12.2 3.8 18.3 5.5 10.2 3 21 3.9 31.6 2.9 11.1-2.6 22.4-4.3 33.8-5.3 4.7 0.2 9.4 1 13.8 2.4 4.6 1.3 8.9 3.6 12.4 6.9 2.5 2.7 4.5 5.8 5.8 9.2 1.9 5.1 3.1 10.4 3.5 15.8 0.2 4.8 0.6 9.6 1.2 14.4 1.7 7.7 5.4 14.9 10.6 20.9 5.3 5.8 11 11.2 17.2 16 5.9 5.2 12.1 10 18.6 14.4 3.1 2.1 6.2 4 9.1 6.3 3 2.2 5.5 5 7.4 8.2 2.4 4.4 3.2 9.5 2 14.4"
              fill="#F5BD0C"
            ></path>
            <path
              d="M887.7 829.8c-2 5.2-4.9 10-8.5 14.3-8.4 9-18.6 16.2-29.8 21.2-19 8.8-37.5 18.6-55.5 29.3-11.7 7.8-22.6 16.6-32.7 26.4-8.3 8.7-17.2 16.7-26.6 24.2-9.8 7.2-21.1 12.1-33.1 14-14.7 1.9-29.6-0.4-43.1-6.5-9.7-3.7-18.1-10.2-24-18.8-5-9.2-7.3-19.5-6.8-29.9 0.6-18.3 2.8-36.5 6.6-54.5 2.6-15 5.2-30 6.8-45.1 2.8-27.6 3.1-55.3 1-82.9-0.5-4.6-0.5-9.3 0-13.9 0.6-9.4 8.5-16.6 18-16.5 4.3-0.1 8.6 0.3 12.8 1.1 10 1.2 20 2.9 29.8 5.2 6.1 1.6 12.2 3.8 18.3 5.5 10.2 3 21 3.9 31.6 2.9 11.1-2.6 22.4-4.3 33.8-5.3 4.7 0.2 9.4 1 13.8 2.4 4.6 1.3 8.9 3.6 12.4 6.9 2.5 2.7 4.5 5.8 5.8 9.2 1.9 5.1 3.1 10.4 3.5 15.8 0.2 4.8 0.6 9.6 1.2 14.4 1.7 7.7 5.4 14.9 10.6 20.9 5.3 5.8 11 11.2 17.2 16 5.9 5.2 12.1 10 18.6 14.4 3.1 2.1 6.2 4 9.1 6.3 3 2.2 5.5 5 7.4 8.2 2.4 4.4 3.2 9.5 2 14.4M259.4 676.3c4.9-1.9 10.2-2.4 15.4-1.4 5.2 1 10.1 3.1 14.4 6.1 8.3 6.3 15.5 14.1 21.2 22.8 14.1 19.4 27.6 39.2 39.9 59.8 10 16.7 19.1 33.9 30.6 49.6 7.5 10.2 16 19.7 23.6 29.9 7.9 10 13.9 21.4 17.6 33.5 4.4 16.1 2.6 33.2-4.9 48.1-5.4 10.4-13.5 19.1-23.4 25.1-10 6-21.5 9-33.2 8.7-18.4-2.5-36.2-8.1-52.6-16.6-34.9-13.9-72.8-18.3-108.8-29.1-11.1-3.3-21.9-7.3-33.1-10.3-5-1.2-9.9-2.7-14.7-4.7-4.7-2-8.8-5.4-11.5-9.7-2-3.5-3-7.5-2.9-11.5 0.1-4 0.9-7.9 2.3-11.5 2.7-7.5 7.1-14.2 10-21.6 4.4-12.2 6.1-25.3 5-38.2-0.6-12.9-2.9-25.8-3.6-38.7-0.6-5.8-0.4-11.6 0.6-17.3 1.5-11.4 10.4-20.5 21.9-22.2 5.3-0.9 10.6-1.3 15.9-1 5.3 0.3 10.7 0.3 16 0 5.3-0.3 10.6-1.8 15.3-4.3 4.3-2.6 8.1-6.2 11-10.4 2.9-4.2 5.5-8.5 7.9-13 2.4-4.5 5.1-8.7 8.3-12.7 3-4.1 7.1-7.2 11.8-9.4"
              fill="#F5BD0C"
            ></path>
            <path
              d="M259.4 676.4c4.9-1.9 10.2-2.4 15.4-1.4 5.2 1 10.1 3.1 14.4 6.1 8.3 6.3 15.5 14.1 21.2 22.8 14.1 19.4 27.6 39.2 39.9 59.8 10 16.7 19.1 33.9 30.6 49.6 7.5 10.2 16 19.7 23.6 29.9 7.9 10 13.9 21.4 17.6 33.5 4.4 16.1 2.6 33.2-4.9 48.1-5.4 10.4-13.5 19.1-23.4 25.1-10 6-21.5 9-33.2 8.7-18.4-2.5-36.2-8.1-52.6-16.6-34.9-13.9-72.8-18.3-108.8-29.1-11.1-3.3-21.9-7.3-33.1-10.3-5-1.2-9.9-2.7-14.7-4.7-4.7-2-8.8-5.4-11.5-9.7-2-3.5-3-7.5-2.9-11.5 0.1-4 0.9-7.9 2.3-11.5 2.7-7.5 7.1-14.2 10-21.6 4.4-12.2 6.1-25.3 5-38.2-0.6-12.9-2.9-25.7-3.6-38.7-0.6-5.8-0.4-11.6 0.6-17.3 1.5-11.4 10.4-20.5 21.9-22.2 5.3-0.9 10.6-1.3 15.9-1 5.3 0.3 10.7 0.3 16 0 5.3-0.3 10.6-1.8 15.3-4.3 4.3-2.6 8.1-6.2 11-10.4 2.9-4.2 5.5-8.5 7.9-13 2.4-4.5 5.1-8.7 8.3-12.7 3-4.1 7.1-7.3 11.8-9.4"
              fill="#F5BD0C"
            ></path>
            <path
              d="M267.1 684.8c4.4-1.7 9.3-2 13.9-0.9s8.9 3.2 12.6 6.2c7.1 6.2 13.1 13.6 17.6 21.9 12 19.4 23.7 39 34.6 59 7.9 15.3 16.8 30.1 26.6 44.2 6.8 9.2 14.6 17.6 21.6 26.6 7.3 8.9 12.8 19 16.2 29.9 4 14.3 2.3 29.6-4.5 42.9-5 9.4-12.5 17.3-21.7 22.6-9.2 5.4-19.8 8-30.4 7.5-16.7-2.6-32.9-7.6-48.2-14.9-30.4-11.1-63.5-12.5-94.7-21.2-11.2-3-22.1-7.1-33.4-9.9-5-1.1-10-2.5-14.8-4.3-4.8-1.8-9-5.2-11.8-9.5-1.8-3.4-2.7-7.2-2.5-11 0.2-3.8 1-7.6 2.4-11.2 2.7-7.1 7-13.6 9.7-20.7 3.8-11 5.1-22.6 3.9-34.2-0.8-11.5-2.9-22.9-3.5-34.5-0.4-5.1-0.2-10.3 0.7-15.4 0.9-5.1 3.3-9.8 6.9-13.6 4.2-3.8 9.4-6.3 15-7 5.6-0.7 11.2-0.7 16.7 0 5.6 0.7 11.2 0.9 16.8 0.8 11 0 21-6.4 25.7-16.4 2.3-4.5 4.3-9.2 5.9-13.9 1.7-4.8 4-9.3 6.7-13.6 2.8-4.3 6.8-7.7 11.5-9.7"
              fill="#F5BD0C"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
            Linux 客户端发布
          </h3>
          <div
            class="tooltip flex items-center text-stone-300"
            data-tip="需服务器安装有git、go、gzip、dpkg-dev、createrepo/createrepo-c"
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
        </header>
        <div
          class="rounded-md border border-stone-200 mt-4 mb-3 gap-2 max-w-sm bg-stone-50 p-2"
        >
          <div class="flex flex-col justify-start">
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              当前源码仓库
            </div>
            <div class="w-full text-left max-w-xl break-all text-gray-500">
              {{
                ClientVersion.linux.url && ClientVersion.linux.url != ""
                  ? ClientVersion.linux.url
                  : "未设置"
              }}
            </div>
            <div class="w-full text-left font-bold max-w-xl text-gray-500">当前版本</div>
            <div class="w-full text-left max-w-xl text-gray-500">
              {{
                ClientVersion.linux.version && ClientVersion.linux.version != ""
                  ? ClientVersion.linux.version
                  : "尚未构建本"
              }}
            </div>
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              最近一次构建情况
            </div>
            <div class="w-full text-left max-w-xl break-all text-gray-500">
              {{
                ClientVersion.linux.buildst && ClientVersion.linux.buildst != ""
                  ? ClientVersion.linux.buildst
                  : "尚未进行"
              }}
            </div>
          </div>
        </div>
        <div class="flex w-full max-w-sm">
          <p class="text-gray-600 max-w-sm min-w-fit">仓库地址</p>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                /*
                (ClientVersion.linux.buildst &&
                  ClientVersion.linux.buildst == '正在进行') ||
                  */
                !isValidURL(wantLinuxRepo)
              "
              @click="publishLinux"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              构建
            </button>
          </div>
        </div>
        <div
          :class="{
            'border-red-500 hover:border-red-700':
              wantLinuxRepo == '' || !isValidURL(wantLinuxRepo),
            'border-stone-200 hover:border-stone-400':
              wantLinuxRepo != '' && isValidURL(wantLinuxRepo),
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="wantLinuxRepo"
          />
        </div>
        <div class="flex w-full max-w-sm">
          <p class="text-gray-600 max-w-sm min-w-fit">
            用户名:口令 (选填, 填 <strong>clear</strong> 清除)
          </p>
        </div>
        <div
          class="mt-1 max-w-sm flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="wantLinuxRepoCred"
          />
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
              :action="uploadURL + '/win'"
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
      <div>
        <header class="max-w-sm flex mt-4">
          <svg
            t="1679468518353"
            viewBox="0 0 1024 1024"
            version="1.1"
            xmlns="http://www.w3.org/2000/svg"
            p-id="1724"
            width="28"
            height="28"
          >
            <path
              d="M645.289723 165.758826C677.460161 122.793797 701.865322 62.036894 693.033384 0c-52.607627 3.797306-114.089859 38.61306-149.972271 84.010072-32.682435 41.130375-59.562245 102.313942-49.066319 161.705521 57.514259 1.834654 116.863172-33.834427 151.294929-79.956767zM938.663644 753.402663c-23.295835 52.820959-34.517089 76.415459-64.511543 123.177795-41.855704 65.279538-100.905952 146.644295-174.121433 147.198957-64.980873 0.725328-81.748754-43.30636-169.982796-42.751697-88.234042 0.46933-106.623245 43.605024-171.732117 42.965029-73.130149-0.682662-129.065752-74.026142-170.964122-139.348347-117.11917-182.441374-129.44975-396.626524-57.172928-510.545717 51.327636-80.895427 132.393729-128.212425 208.553189-128.212425 77.482118 0 126.207106 43.519692 190.377318 43.519692 62.292892 0 100.137957-43.647691 189.779989-43.647691 67.839519 0 139.732344 37.802399 190.889315 103.03927-167.678812 94.036667-140.543004 339.069598 28.885128 404.605134z"
              fill="#0B0B0A"
              p-id="1725"
            ></path>
          </svg>
          <h3 class="text-xl font-semibold tracking-tight ml-4 min-w-fit">
            iOS 客户端发布
          </h3>
        </header>
        <div
          class="rounded-md border border-stone-200 mt-4 mb-3 gap-2 max-w-sm bg-stone-50 p-2"
        >
          <div class="flex flex-col justify-start">
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              当前AppStore版本
            </div>
            <div class="w-full text-left max-w-xl text-gray-500">
              {{
                ClientVersion.ios_store.version && ClientVersion.ios_store.version != ""
                  ? ClientVersion.ios_store.version
                  : "未设置"
              }}
            </div>
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              当前AppStore地址
            </div>
            <div class="w-full text-left max-w-xl break-all text-gray-500">
              {{
                ClientVersion.ios_store.url && ClientVersion.ios_store.url != ""
                  ? ClientVersion.ios_store.url
                  : "未设置"
              }}
            </div>
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              当前TestFlight版本
            </div>
            <div class="w-full text-left max-w-xl text-gray-500">
              {{
                ClientVersion.ios_test.version && ClientVersion.ios_test.version != ""
                  ? ClientVersion.ios_test.version
                  : "未设置"
              }}
            </div>
            <div class="w-full text-left font-bold max-w-xl text-gray-500">
              当前TestFlight地址
            </div>
            <div class="w-full text-left max-w-xl break-all text-gray-500">
              {{
                ClientVersion.ios_test.url && ClientVersion.ios_test.url != ""
                  ? ClientVersion.ios_test.url
                  : "未设置"
              }}
            </div>
          </div>
        </div>
        <div class="flex w-full max-w-sm">
          <p class="text-gray-600 max-w-sm min-w-fit">AppStore版本号</p>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                wantIOSStoreVersion.version == '' ||
                wantIOSStoreVersion.url == '' ||
                !isValidURL(wantIOSStoreVersion.url) ||
                (wantIOSStoreVersion.version == ClientVersion.ios_store.version &&
                  wantIOSStoreVersion.url == ClientVersion.ios_store.url)
              "
              @click="publishIOSToStore"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              发布
            </button>
          </div>
        </div>
        <div
          :class="{
            'border-red-500 hover:border-red-700': wantIOSStoreVersion.version == '',
            'border-stone-200 hover:border-stone-400': wantIOSStoreVersion.version != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="wantIOSStoreVersion.version"
          />
        </div>

        <div class="flex flex-row w-full max-w-sm justify-start mt-2 space-x-2">
          <p class="text-gray-600 min-w-fit">AppStore URL</p>
        </div>
        <div
          class="max-w-sm flex w-full border rounded-md relative overflow-hidden min-w-0"
          :class="{
            'border-red-500 hover:border-red-700': !isValidURL(wantIOSStoreVersion.url),
            'border-stone-200 hover:border-stone-400': isValidURL(
              wantIOSStoreVersion.url
            ),
          }"
        >
          <input
            class="outline-none py-2 px-3 w-full h-9 font-mono text-sm text-ellipsis"
            v-model="wantIOSStoreVersion.url"
          />
        </div>
        <div class="flex w-full max-w-sm mt-1">
          <p class="text-gray-600 max-w-sm min-w-fit">TestFlight版本号</p>
          <div class="w-full flex justify-end">
            <button
              :disabled="
                wantIOSTestVersion.version == '' ||
                wantIOSTestVersion.url == '' ||
                !isValidURL(wantIOSTestVersion.url) ||
                (wantIOSTestVersion.version == ClientVersion.ios_test.version &&
                  wantIOSTestVersion.url == ClientVersion.ios_test.url)
              "
              @click="publishIOSToTestflight"
              class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-7 min-h-fit"
            >
              发布
            </button>
          </div>
        </div>
        <div
          :class="{
            'border-red-500 hover:border-red-700': wantIOSTestVersion.version == '',
            'border-stone-200 hover:border-stone-400': wantIOSTestVersion.version != '',
          }"
          class="mt-1 max-w-sm flex border rounded-md relative overflow-hidden min-w-0"
        >
          <input
            class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
            v-model="wantIOSTestVersion.version"
          />
        </div>

        <div class="flex flex-row w-full max-w-sm justify-start mt-2 space-x-2">
          <p class="text-gray-600 min-w-fit">TestFlight URL</p>
        </div>
        <div
          class="max-w-sm flex w-full border rounded-md relative overflow-hidden min-w-0"
          :class="{
            'border-red-500 hover:border-red-700': !isValidURL(wantIOSTestVersion.url),
            'border-stone-200 hover:border-stone-400': isValidURL(wantIOSTestVersion.url),
          }"
        >
          <input
            class="outline-none py-2 px-3 w-full h-9 font-mono text-sm text-ellipsis"
            v-model="wantIOSTestVersion.url"
          />
        </div>
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
