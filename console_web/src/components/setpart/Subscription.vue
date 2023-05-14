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

const billingUsage = ref({});

onMounted(() => {
  axios
    .get("/admin/api/subscription")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        billingUsage.value = response.data["data"]["billingUsage"];
      } else {
        toastMsg.value = "出错了：" + response.data["status"].substring(6);
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      toastMsg.value = "出错了：" + error;
      toastShow.value = true;
    });
});
</script>

<template>
  <div class="flex-1">
    <div
      class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center"
    >
      <h1 class="mr-2" tabindex="-1">账单</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>查看管理您的网络账单计划</p>
    </div>
    <div class="mt-10">
      <div class="space-y-10">
        <div>
          <div
            class="flex flex-wrap gap-y-2 justify-between items-center border-b pb-3 mb-4"
          >
            <h2 class="text-xl font-medium min-w-[33%]">你的计划</h2>
            <div>
              <button
                disabled
                class="btn bg-blue-500 border-blue-500 text-white enabled:hover:bg-blue-600 enabled:hover:border-blue-600 disabled:text-blue-50 disabled:bg-blue-300 disabled:border-blue-300 px-3 text-sm py-[0.35rem] ml-2 h-9 min-h-fit"
              >
                <span class="flex-1">升级...</span>
              </button>
            </div>
          </div>
          <div class="flex flex-wrap gap-x-6 gap-y-4">
            <div class="flex flex-1 gap-x-3 items-center min-w-[50%]">
              <svg
                width="64"
                height="64"
                viewBox="0 0 60 60"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect x="23" y="7" width="30" height="30" fill="#C65835"></rect>
                <circle cx="23" cy="36" r="16" fill="#496495"></circle>
              </svg>
              <div class="flex flex-1 justify-center flex-col">
                <div class="flex items-center flex-wrap gap-x-2">
                  <div class="text-xl font-medium whitespace-nowrap">开源无限量</div>
                </div>
                <p class="text-gray-600 text-sm max-w-md">￥0 /用户·月</p>
              </div>
            </div>
            <div class="flex items-center gap-4">
              <div class="min-w-[6rem]">
                <div class="uppercase tracking-wide text-xs text-gray-500">月账单</div>
                <div class="text-xl font-medium">￥0.00</div>
              </div>
              <div class="min-w-[6rem]">
                <div class="uppercase tracking-wide text-xs text-gray-500">下次支付</div>
                <div class="text-xl font-medium">猴年 马月</div>
              </div>
            </div>
          </div>
        </div>
        <div>
          <div
            class="flex flex-wrap gap-y-2 justify-between items-center border-b pb-3 mb-4"
          >
            <h2 class="text-xl font-medium min-w-[33%]">用量</h2>
          </div>
          <table>
            <tbody>
              <tr>
                <td class="align-top pr-8 pb-2 whitespace-nowrap">
                  <span class="font-mono">{{ billingUsage.usage.users }}</span
                  ><span class="px-2 text-gray-500 break-all">共</span
                  ><span>{{
                    billingUsage.allowance.users.total.unlimited
                      ? "∞"
                      : billingUsage.allowance.users.total.amount
                  }}</span>
                </td>
                <td class="align-top pb-2">用户</td>
              </tr>
              <tr>
                <td class="align-top pr-8 pb-2 whitespace-nowrap">
                  <span class="font-mono">{{ billingUsage.usage.admin_users }}</span
                  ><span class="px-2 text-gray-500 break-all">共</span
                  ><span>{{
                    billingUsage.allowance.admin_users.total.unlimited
                      ? "∞"
                      : billingUsage.allowance.admin_users.total.amount
                  }}</span>
                </td>
                <td class="align-top pb-2">管理员用户</td>
              </tr>
              <tr>
                <td class="align-top pr-8 pb-2 whitespace-nowrap">
                  <span class="font-mono">{{ billingUsage.usage.acl_named_users }}</span
                  ><span class="px-2 text-gray-500 break-all">共</span
                  ><span>{{
                    billingUsage.allowance.acl_named_users.total.unlimited
                      ? "∞"
                      : billingUsage.allowance.acl_named_users.total.amount
                  }}</span>
                </td>
                <td class="align-top pb-2">ACL独立控制用户</td>
              </tr>
              <tr>
                <td class="align-top pr-8 pb-2 whitespace-nowrap">
                  <span class="font-mono">{{ billingUsage.usage.devices }}</span
                  ><span class="px-2 text-gray-500 break-all">共</span
                  ><span>{{
                    billingUsage.allowance.devices.total.unlimited
                      ? "∞"
                      : billingUsage.allowance.devices.total.amount
                  }}</span>
                </td>
                <td class="align-top pb-2">设备</td>
              </tr>
              <tr>
                <td class="align-top pr-8 pb-2 whitespace-nowrap">
                  <span class="font-mono">{{ billingUsage.usage.subnets }}</span
                  ><span class="px-2 text-gray-500 break-all">共</span
                  ><span>{{
                    billingUsage.allowance.subnets.total.unlimited
                      ? "∞"
                      : billingUsage.allowance.subnets.total.amount
                  }}</span>
                </td>
                <td class="align-top pb-2">子网转发路由</td>
              </tr>
            </tbody>
          </table>
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
