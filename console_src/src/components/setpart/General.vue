<script setup>
import { onMounted, ref, watch } from "vue";
import Toast from "../Toast.vue";

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

const UserAccount = ref("");
const Basedomain = ref("");
const UserName = ref("");
const UserNameHead = ref("");
const OrgName = ref("");

//服务器设置的最大密钥过期时长
const MaxKeyExpiry = ref(180);
//输入框设置的密钥过期时长
const keyExpiryInputValue = ref(180);
const keyExpirySubDis = ref(false);
const keyExpiryAddDis = ref(false);
const keyExpirySaveDis = ref(false);

const copyBtnText = ref("复制");

function copyOrgName() {
  navigator.clipboard.writeText(OrgName.value).then(function () {
    copyBtnText.value = "已复制!";
    setTimeout(() => {
      copyBtnText.value = "复制";
    }, 3000);
  });
}

function updateKeyExpiryBtns() {
  if (Number(keyExpiryInputValue.value) > 1) {
    keyExpirySubDis.value = false;
  } else {
    keyExpirySubDis.value = true;
  }
  if (Number(keyExpiryInputValue.value) < 180) {
    keyExpiryAddDis.value = false;
  } else {
    keyExpiryAddDis.value = true;
  }
  if (Number(keyExpiryInputValue.value) != MaxKeyExpiry.value) {
    keyExpirySaveDis.value = false;
  } else {
    keyExpirySaveDis.value = true;
  }
}
function keyExpiryCheck(isChange) {
  keyExpiryInputValue.value = keyExpiryInputValue.value
    .replace(/[^\d]+/g, "")
    .replace(/^0+(\d)/, "$1");
  if (keyExpiryInputValue.value == "") keyExpiryInputValue.value = 0;
  if (isChange) {
    if (Number(keyExpiryInputValue.value) > 180) keyExpiryInputValue.value = 180;
  }
  updateKeyExpiryBtns();
}
function keyExpiryChange(isAdd) {
  if (isAdd == true) {
    keyExpiryInputValue.value = Number(keyExpiryInputValue.value) + 1;
  } else {
    keyExpiryInputValue.value = Number(keyExpiryInputValue.value) - 1;
  }
  updateKeyExpiryBtns();
}
function resetKeyExpiryInput() {
  keyExpiryInputValue.value = MaxKeyExpiry.value;
  updateKeyExpiryBtns();
}

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

  axios
    .get("/admin/api/netsettings")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        MaxKeyExpiry.value = response.data["data"]["maxKeyDurationDays"];
        keyExpiryInputValue.value = response.data["data"]["maxKeyDurationDays"];
        updateKeyExpiryBtns();
      } else {
        if (response.data["status"].substring(6) == "用户信息核对失败") {
          //TODO:token失效跳转
        }
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

//服务端请求
function updateKeyExpiry() {

  axios
    .post("/admin/api/netsetting/updatekeyexpiry", {
      maxKeyDurationDays: Number(keyExpiryInputValue.value),
    })
    .then(function (response) {
      if (response.data["status"] == "success") {
        MaxKeyExpiry.value = response.data["data"];
        updateKeyExpiryBtns();
        toastMsg.value = "已更新您网络中节点密钥最长有效期！";
        toastShow.value = true;
      } else {
        toastMsg.value = "失败：" + response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      console.log(error);
    });
}
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
                  @click="copyOrgName"
                  class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap"
                >
                  {{ copyBtnText }}
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
              <div class="relative">
                <input
                  v-model="keyExpiryInputValue"
                  @input="keyExpiryCheck(false)"
                  @blur="keyExpiryCheck(true)"
                  class="input z-0 border focus:outline-blue-500/60 hover:border border-stone-200 hover:border-stone-400 rounded-r-none h-9 min-h-fit"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  id="key-expiry-duration"
                  tabindex="0"
                />
                <div
                  class="bg-white top-1 bottom-1 right-1 rounded-r-md absolute flex items-center"
                >
                  <div class="flex items-center">
                    <button
                      @click="keyExpiryChange(false)"
                      class="btn btn-ghost btn-sm px-2 hover:bg-stone-100 disabled:bg-transparent"
                      :disabled="keyExpirySubDis"
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
                      @click="keyExpiryChange(true)"
                      class="btn btn-ghost btn-sm px-2 hover:bg-stone-100 disabled:bg-transparent"
                      :disabled="keyExpiryAddDis"
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
                @click="updateKeyExpiry"
                class="btn border-0 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/60 text-white disabled:text-white/60 h-9 min-h-fit"
                :disabled="keyExpirySaveDis"
              >
                保存
              </button>
              <button
                @click="resetKeyExpiryInput"
                class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3"
                :disabled="keyExpirySaveDis"
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
            :disabled="devMode"
            class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit"
          >
            注销蜃境…
          </button>
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
