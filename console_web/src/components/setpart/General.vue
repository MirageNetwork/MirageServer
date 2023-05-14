<script setup>
import { onMounted, ref, watch } from "vue";
import Toast from "../Toast.vue";

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});
const OrgName = ref("");

const copyBtnText = ref("复制");

function copyOrgName() {
  navigator.clipboard.writeText(OrgName.value).then(function () {
    copyBtnText.value = "已复制!";
    setTimeout(() => {
      copyBtnText.value = "复制";
    }, 3000);
  });
}

onMounted(() => {
  axios
    .get("/admin/api/self")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        OrgName.value = response.data["data"]["orgname"];
      } else {
        toastMsg.value = "获取机构名称失败:" + response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      toastMsg.value = "获取机构名称失败:" + error;
      toastShow.value = true;
    });
});
</script>

<template>
  <div class="flex-1">
    <div
      class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center"
    >
      <h1 class="mr-2" tabindex="-1">通用</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>管理您的蜃境网络配置</p>
    </div>
    <div class="mt-10">
      <div class="space-y-10">
        <div>
          <header class="max-w-2xl">
            <h3 class="text-xl font-semibold tracking-tight">机构</h3>
            <p class="mt-1 text-gray-600">这是您所属机构名称，无法修改！</p>
          </header>
          <div class="mt-4">
            <div class="max-w-sm">
              <div
                class="flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
              >
                <input
                  onclick="this.select()"
                  class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
                  readonly
                  :value="OrgName"
                />
                <button
                  @click="copyOrgName"
                  class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
                >
                  {{ copyBtnText }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="mt-8">
        <header class="max-w-2xl">
          <h3 class="text-xl font-semibold tracking-tight">注销蜃境</h3>
          <p class="mt-1 text-gray-600">
            永久注销蜃境网络（删除所有设备及用户信息），无法恢复！
          </p>
        </header>
        <div class="mt-4">
          <div
            class="flex items-center rounded-md px-4 py-2 border border-stone-200 bg-stone-50 font-light text-gray-600 h-9 min-h-fit w-fit"
          >
            <div class="pt-px mr-2">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="1.125em"
                height="1.125em"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="16" x2="12" y2="12"></line>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
              </svg>
            </div>
            网络所有者必须联系支持
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

<style scoped></style>
