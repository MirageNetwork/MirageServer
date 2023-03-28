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

//服务器设置的最大密钥过期时长
const MaxKeyExpiry = ref(180);
//输入框设置的密钥过期时长
const keyExpiryInputValue = ref(180);
const keyExpirySubDis = ref(false);
const keyExpiryAddDis = ref(false);
const keyExpirySaveDis = ref(false);

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
  //if (keyExpiryInputValue.value == "") keyExpiryInputValue.value = 1;
  if (isChange) {
    if (keyExpiryInputValue.value == "") keyExpiryInputValue.value = MaxKeyExpiry.value;
    if (Number(keyExpiryInputValue.value) == 0) keyExpiryInputValue.value = 1;
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
    .get("/admin/api/netsettings")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        MaxKeyExpiry.value = response.data["data"]["maxKeyDurationDays"];
        keyExpiryInputValue.value = response.data["data"]["maxKeyDurationDays"];
        updateKeyExpiryBtns();
      } else {
        toastMsg.value = "失败：" + response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      toastMsg.value = "失败：" + error;
      toastShow.value = true;
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
      toastMsg.value = "失败：" + error;
      toastShow.value = true;
    });
}
</script>

<template>
  <div class="flex-1">
    <div
      class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center"
    >
      <h1 class="mr-2" tabindex="-1">设备</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>管理您的蜃境网络设备配置</p>
    </div>
    <div class="mt-10">
      <div class="space-y-10">
        <div>
          <header class="max-w-2xl">
            <h3 class="text-xl font-semibold tracking-tight">设备批准</h3>
            <p class="mt-1 text-gray-600">要求新添加设备在访问网络前需要被管理员批准。</p>
          </header>
          <div class="mt-4">
            <span data-state="delayed-open"
              ><div class="flex items-center">
                <input
                  disabled
                  id="require-approval"
                  type="checkbox"
                  class="toggle mr-3"
                /><label class="font-medium cursor-pointer" for="require-approval"
                  >手动批准新设备</label
                >
              </div></span
            >
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
