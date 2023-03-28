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

const orgProvider = ref("");

onMounted(() => {
  axios
    .get("/admin/api/netsettings")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        orgProvider.value = response.data["data"]["provider"];
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
      <h1 class="mr-2" tabindex="-1">用户</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>管理您的蜃境网络用户配置</p>
    </div>
    <div class="mt-10">
      <div class="space-y-10">
        <div>
          <header class="max-w-2xl">
            <h3 class="text-xl font-semibold tracking-tight">身份提供商</h3>
            <p class="mt-1 text-gray-600">
              你的用户如何被鉴别登入蜃境网络的。
              <a
                :href="'mailto:gps949@nopkt.com?subject=[修改我的身份提供商]'"
                class="link"
                target="_blank"
                rel="noopener noreferrer"
                >联系支持</a
              >
              来修改。
            </p>
          </header>
          <div class="mt-4">
            <div
              class="rounded border border-stone-300 bg-stone-50 py-2 px-4 inline-flex items-center font-medium"
            >
              <span class="mr-2"
                ><svg
                  v-if="orgProvider == 'Microsoft'"
                  width="17"
                  height="17"
                  viewBox="0 0 16 16"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path d="M0 0H7.57886V7.57886H0V0Z" fill="#F25022"></path>
                  <path d="M0 8.42114H7.57886V16H0V8.42114Z" fill="#00A4EF"></path>
                  <path d="M8.42114 0H16V7.57886H8.42114V0Z" fill="#7FBA00"></path>
                  <path d="M8.42114 8.42114H16V16H8.42114V8.42114Z" fill="#FFB900"></path>
                </svg>
                <svg
                  v-if="orgProvider == 'Github'"
                  t="1679387527759"
                  class="icon ml-8 mr-3"
                  viewBox="0 0 1024 1024"
                  version="1.1"
                  xmlns="http://www.w3.org/2000/svg"
                  p-id="3364"
                  width="17"
                  height="17"
                >
                  <path
                    d="M0 524.714667c0 223.36 143.146667 413.269333 342.656 482.986666 26.88 6.826667 22.784-12.373333 22.784-25.344v-88.618666c-155.136 18.176-161.322667-84.48-171.818667-101.589334-21.077333-35.968-70.741333-45.141333-55.936-62.250666 35.328-18.176 71.338667 4.608 112.981334 66.261333 30.165333 44.672 89.002667 37.12 118.912 29.653333a144.64 144.64 0 0 1 39.68-69.546666c-160.682667-28.757333-227.712-126.848-227.712-243.541334 0-56.576 18.688-108.586667 55.253333-150.570666-23.296-69.205333 2.176-128.384 5.546667-137.173334 66.474667-5.973333 135.424 47.573333 140.8 51.754667 37.76-10.197333 80.810667-15.573333 128.981333-15.573333 48.426667 0 91.733333 5.546667 129.706667 15.872 12.8-9.813333 76.885333-55.765333 138.666666-50.133334 3.285333 8.789333 28.16 66.602667 6.272 134.826667 37.077333 42.069333 55.936 94.549333 55.936 151.296 0 116.864-67.413333 215.04-228.565333 243.456a145.92 145.92 0 0 1 43.52 104.106667v128.64c0.896 10.282667 0 20.48 17.194667 20.48 202.410667-68.224 348.16-259.541333 348.16-484.906667C1023.018667 242.176 793.941333 13.312 511.573333 13.312 228.864 13.184 0 242.090667 0 524.714667z"
                    fill="#000000"
                    p-id="3365"
                  ></path>
                </svg>
                <svg
                  v-if="orgProvider == 'Google'"
                  t="1679449475826"
                  class="icon ml-8 mr-3"
                  viewBox="0 0 1024 1024"
                  version="1.1"
                  xmlns="http://www.w3.org/2000/svg"
                  p-id="4669"
                  width="17"
                  height="17"
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
                <svg
                  v-if="orgProvider == 'Apple'"
                  t="1679468518353"
                  class="icon ml-8 mr-3"
                  viewBox="0 0 1024 1024"
                  version="1.1"
                  xmlns="http://www.w3.org/2000/svg"
                  p-id="1724"
                  width="17"
                  height="17"
                >
                  <path
                    d="M645.289723 165.758826C677.460161 122.793797 701.865322 62.036894 693.033384 0c-52.607627 3.797306-114.089859 38.61306-149.972271 84.010072-32.682435 41.130375-59.562245 102.313942-49.066319 161.705521 57.514259 1.834654 116.863172-33.834427 151.294929-79.956767zM938.663644 753.402663c-23.295835 52.820959-34.517089 76.415459-64.511543 123.177795-41.855704 65.279538-100.905952 146.644295-174.121433 147.198957-64.980873 0.725328-81.748754-43.30636-169.982796-42.751697-88.234042 0.46933-106.623245 43.605024-171.732117 42.965029-73.130149-0.682662-129.065752-74.026142-170.964122-139.348347-117.11917-182.441374-129.44975-396.626524-57.172928-510.545717 51.327636-80.895427 132.393729-128.212425 208.553189-128.212425 77.482118 0 126.207106 43.519692 190.377318 43.519692 62.292892 0 100.137957-43.647691 189.779989-43.647691 67.839519 0 139.732344 37.802399 190.889315 103.03927-167.678812 94.036667-140.543004 339.069598 28.885128 404.605134z"
                    fill="#0B0B0A"
                    p-id="1725"
                  ></path>
                </svg>
                <svg
                  v-if="orgProvider == 'WXScan'"
                  xmlns="http://www.w3.org/2000/svg"
                  class="ml-8 mr-3"
                  viewBox="0 0 24 24"
                  width="17"
                  height="17"
                >
                  <path fill="none" d="M0 0h24v24H0z" />
                  <path
                    d="M15.84 12.691l-.067.02a1.522 1.522 0 0 1-.414.062c-.61 0-.954-.412-.77-.921.136-.372.491-.686.925-.831.672-.245 1.142-.804 1.142-1.455 0-.877-.853-1.587-1.905-1.587s-1.904.71-1.904 1.587v4.868c0 1.17-.679 2.197-1.694 2.778a3.829 3.829 0 0 1-1.904.502c-1.984 0-3.598-1.471-3.598-3.28 0-.576.164-1.117.451-1.587.444-.73 1.184-1.287 2.07-1.541a1.55 1.55 0 0 1 .46-.073c.612 0 .958.414.773.924-.126.347-.466.645-.861.803a2.162 2.162 0 0 0-.139.052c-.628.26-1.061.798-1.061 1.422 0 .877.853 1.587 1.905 1.587s1.904-.71 1.904-1.587V9.566c0-1.17.679-2.197 1.694-2.778a3.829 3.829 0 0 1 1.904-.502c1.984 0 3.598 1.471 3.598 3.28 0 .576-.164 1.117-.451 1.587-.442.726-1.178 1.282-2.058 1.538zM2 12c0 5.523 4.477 10 10 10s10-4.477 10-10S17.523 2 12 2 2 6.477 2 12z"
                    fill="rgba(56,186,109,1)"
                  />
                </svg>
              </span>

              {{ orgProvider == "WXScan" ? "微信小程序" : orgProvider }}
            </div>
          </div>
        </div>
        <div>
          <header class="max-w-2xl">
            <h3 class="text-xl font-semibold tracking-tight">用户批准</h3>
            <p class="mt-1 text-gray-600">要求新用户在访问网络前需要被管理员批准。</p>
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
                  >手动批准新用户</label
                >
              </div></span
            >
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
