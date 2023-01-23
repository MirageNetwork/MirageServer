<script setup>
import { onMounted, ref } from "vue";

const UserAccount = ref("");
const Basedomain = ref("");
const UserName = ref("");
const UserNameHead = ref("");
const OrgName = ref("");

onMounted(() => {
  axios
    .get("/admin/api/self")
    .then(function (response) {
      // 处理成功情况

      if (response.data["errormsg"] == undefined || response.data["errormsg"] === "") {
        UserAccount.value = response.data["useraccount"];
        Basedomain.value = response.data["basedomain"];
        UserName.value = response.data["username"];
        UserNameHead.value = response.data["usernamehead"];
        OrgName.value = response.data["orgname"];
      }
    })
    .catch(function (error) {
      // 处理错误情况
      alert(error);
    })
    .then(function () {
      // 总是会执行
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
            <p class="mt-1 text-gray-600">
              这是您所属机构名称（个人用户显示您的账户），无法修改！
            </p>
          </header>
          <div class="mt-4">
            <div class="max-w-sm">
              <div
                class="flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0"
              >
                <input
                  class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
                  readonly
                  :value="OrgName"
                />
                <button
                  class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
                >
                  复制
                </button>
              </div>
            </div>
          </div>
        </div>

        <div>
          <header class="max-w-2xl">
            <h3 class="text-xl font-semibold tracking-tight">密钥过期</h3>
            <p class="mt-1 text-gray-600">
              设置设备可以保持登录蜃境网络而不需重新登录认证的天数
            </p>
          </header>
          <div class="mt-4">
            <div class="flex">
              <div class="relative ">
                <input
                  class="input z-10 border focus:outline-blue-500/60 hover:border border-stone-200 hover:border-stone-400 rounded-r-none h-9 min-h-fit"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  id="key-expiry-duration"
                  value="180"
                />
                <div
                  class="bg-white top-1 bottom-1 right-1 rounded-r-md absolute flex items-center"
                >
                  <div class="flex items-center">
                    <button
                      class="btn btn-ghost btn-sm px-2 hover:bg-stone-100"
                      :disabled="false"
                      type="button"
                      tabindex="-1"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <line x1="5" y1="12" x2="19" y2="12"></line>
                      </svg></button
                    ><button
                      class="btn btn-ghost btn-sm px-2 hover:bg-stone-100"
                      :disabled="false"
                      type="button"
                      tabindex="-1"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <line x1="12" y1="5" x2="12" y2="19"></line>
                        <line x1="5" y1="12" x2="19" y2="12"></line>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
              <div
                class="flex items-center px-3 bg-gray-50 text-gray-500 border rounded-r border-l-0 border-gray-300"
              >
                天
              </div>
            </div>
            <p class="text-sm text-gray-500 mt-1">请设置为1~180天</p>
            <div class="mt-4">
              <button
                class="btn border-0 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/60 text-white disabled:text-white/60 h-9 min-h-fit"
                :disabled="true"
              >
                保存
              </button>
              <button
                class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3"
                :disabled="true"
              >
                重置
              </button>
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
          <button
            class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit"
          >
            注销蜃境…
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
