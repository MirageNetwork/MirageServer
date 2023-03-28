<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch, watchEffect } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate } from "vue-router";
import TenantMenu from "./components/TenantMenu.vue";
import RemoveTenant from "./umenu/RemoveTenant.vue";
import Toast from "./components/Toast.vue";

//与框架交互部分

//界面控制部分
const activeBtn = ref(null);
const btnLeft = ref(0);
const btnTop = ref(0);
function refreshTenantMenuPos() {
  if (activeBtn.value != null) {
    btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14;
    btnTop.value = activeBtn.value?.getBoundingClientRect().top;
  }
}
function openTenantMenu(u, event) {
  activeBtn.value = event.target;
  while (activeBtn.value?.tagName != "DIV" && activeBtn.value?.tagName != "div") {
    activeBtn.value = activeBtn.value.parentNode;
  }
  selectTenant.value = u;
  btnLeft.value = activeBtn.value?.getBoundingClientRect().left + 14;
  btnTop.value = activeBtn.value?.getBoundingClientRect().top;
  tenantMenuShow.value = true;
}
function closeTenantMenu() {
  activeBtn.value = null;
  tenantMenuShow.value = false;
}

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

const selectTenant = ref({});
function mouseOnTenant(u) {
  selectTenant.value = u;
  tenantBtnShow.value = true;
}
function mouseLeaveTenant() {
  tenantBtnShow.value = false;
}

const tenantMenuShow = ref(false);
const tenantBtnShow = ref(false);

const removeTenantShow = ref(false);
function showRemoveTenant() {
  tenantBtnShow.value = false;
  closeTenantMenu();

  removeTenantShow.value = true;
}

function doRemoveTenant() {
  axios
    .post("/cockpit/api/tenants", {
      tenantID: selectTenant.value["id"],
      action: "delete_tenant",
    })
    .then(function (response) {
      if (response.data["status"] != "success") {
        toastMsg.value = response.data["status"].substring(6);
        toastShow.value = true;
      } else {
        removeTenantShow.value = false;
        toastMsg.value = "已删除 " + selectTenant.value["name"];
        toastShow.value = true;
        getTenants().then().catch();
      }
    })
    .catch(function (error) {
      toastMsg.value = error;
      toastShow.value = true;
    });
}

//数据填充控制部分
const TenantList = ref({});
const TenantNum = computed(() => {
  return TenantList.value.length;
});
let getTenantsIntID;
function getTenants() {
  return new Promise((resolve, reject) => {
    axios
      .get("/cockpit/api/tenants")
      .then(function (response) {
        if (response.data["status"] != "success") {
          toastMsg.value = "获租户信息出错：" + response.data["status"].substring(6);
          toastShow.value = true;
          reject();
        }

        // 处理成功情况
        TenantList.value = response.data["data"]["tenants"];
        resolve();
      })
      .catch(function (error) {
        // 处理错误情况
        toastMsg.value = "获取用户信息出错：" + error;
        toastShow.value = true;
        reject();
      });
  });
}
onMounted(() => {
  refreshTenantMenuPos();
  window.addEventListener("resize", refreshTenantMenuPos);
  window.addEventListener("scroll", refreshTenantMenuPos);

  getTenants().then().catch();
  getTenantsIntID = setInterval(() => {
    getTenants().then().catch();
  }, 20000);
});

onUnmounted(() => {
  window.removeEventListener("resize", refreshTenantMenuPos);
  window.removeEventListener("scroll", refreshTenantMenuPos);
});

onBeforeRouteLeave(() => {
  clearInterval(getTenantsIntID);
});
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="mb-24">
      <header class="mb-8">
        <div class="flex justify-between items-center">
          <div class="flex items-center">
            <h1 class="text-3xl font-semibold tracking-tight leading-tight mb-2">租户</h1>
          </div>
        </div>
      </header>

      <div
        class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-full px-2 py-1 leading-none text-sm mb-8"
      >
        {{ TenantNum }} 个租户
      </div>
      <table class="table w-full">
        <thead>
          <tr>
            <th class="flex-auto table-cell items-center">租户</th>
            <th class="table-cell items-center md:w-1/4 lg:w-1/5">所有者</th>
            <th class="hidden lg:table-cell items-center lg:w-1/5">创建日期</th>
            <th class="hidden lg:table-cell items-center lg:w-1/5">最近连线</th>
            <th class="table-cell justify-end ml-auto md:ml-0 relative items-center w-8">
              <span class="sr-only">租户操作菜单</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(u, id) in TenantList">
            <tr
              :id="id"
              :v-if="u != nil"
              @mouseenter="mouseOnTenant(u)"
              @mouseleave="mouseLeaveTenant()"
              class="w-full px-0.5 hover"
            >
              <td class="flex-auto flex items-center">
                <div
                  class="relative shrink-0 overflow-hidden transition-all duration-300 w-8 h-8 md:w-10 md:h-10 md:text-xl mr-3"
                >
                  <svg
                    v-if="u.provider == 'Microsoft'"
                    class="w-full h-full"
                    viewBox="0 0 16 16"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path d="M0 0H7.57886V7.57886H0V0Z" fill="#F25022"></path>
                    <path d="M0 8.42114H7.57886V16H0V8.42114Z" fill="#00A4EF"></path>
                    <path d="M8.42114 0H16V7.57886H8.42114V0Z" fill="#7FBA00"></path>
                    <path
                      d="M8.42114 8.42114H16V16H8.42114V8.42114Z"
                      fill="#FFB900"
                    ></path>
                  </svg>
                  <svg
                    v-if="u.provider == 'Github'"
                    class="w-full h-full"
                    t="1679387527759"
                    viewBox="0 0 1024 1024"
                    version="1.1"
                    xmlns="http://www.w3.org/2000/svg"
                    p-id="3364"
                  >
                    <path
                      d="M0 524.714667c0 223.36 143.146667 413.269333 342.656 482.986666 26.88 6.826667 22.784-12.373333 22.784-25.344v-88.618666c-155.136 18.176-161.322667-84.48-171.818667-101.589334-21.077333-35.968-70.741333-45.141333-55.936-62.250666 35.328-18.176 71.338667 4.608 112.981334 66.261333 30.165333 44.672 89.002667 37.12 118.912 29.653333a144.64 144.64 0 0 1 39.68-69.546666c-160.682667-28.757333-227.712-126.848-227.712-243.541334 0-56.576 18.688-108.586667 55.253333-150.570666-23.296-69.205333 2.176-128.384 5.546667-137.173334 66.474667-5.973333 135.424 47.573333 140.8 51.754667 37.76-10.197333 80.810667-15.573333 128.981333-15.573333 48.426667 0 91.733333 5.546667 129.706667 15.872 12.8-9.813333 76.885333-55.765333 138.666666-50.133334 3.285333 8.789333 28.16 66.602667 6.272 134.826667 37.077333 42.069333 55.936 94.549333 55.936 151.296 0 116.864-67.413333 215.04-228.565333 243.456a145.92 145.92 0 0 1 43.52 104.106667v128.64c0.896 10.282667 0 20.48 17.194667 20.48 202.410667-68.224 348.16-259.541333 348.16-484.906667C1023.018667 242.176 793.941333 13.312 511.573333 13.312 228.864 13.184 0 242.090667 0 524.714667z"
                      fill="#000000"
                      p-id="3365"
                    ></path>
                  </svg>
                  <svg
                    v-if="u.provider == 'Google'"
                    class="w-full h-full"
                    t="1679449475826"
                    viewBox="0 0 1024 1024"
                    version="1.1"
                    xmlns="http://www.w3.org/2000/svg"
                    p-id="4669"
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
                    v-if="u.provider == 'Apple'"
                    class="w-full h-full"
                    t="1679468518353"
                    viewBox="0 0 1024 1024"
                    version="1.1"
                    xmlns="http://www.w3.org/2000/svg"
                    p-id="1724"
                  >
                    <path
                      d="M645.289723 165.758826C677.460161 122.793797 701.865322 62.036894 693.033384 0c-52.607627 3.797306-114.089859 38.61306-149.972271 84.010072-32.682435 41.130375-59.562245 102.313942-49.066319 161.705521 57.514259 1.834654 116.863172-33.834427 151.294929-79.956767zM938.663644 753.402663c-23.295835 52.820959-34.517089 76.415459-64.511543 123.177795-41.855704 65.279538-100.905952 146.644295-174.121433 147.198957-64.980873 0.725328-81.748754-43.30636-169.982796-42.751697-88.234042 0.46933-106.623245 43.605024-171.732117 42.965029-73.130149-0.682662-129.065752-74.026142-170.964122-139.348347-117.11917-182.441374-129.44975-396.626524-57.172928-510.545717 51.327636-80.895427 132.393729-128.212425 208.553189-128.212425 77.482118 0 126.207106 43.519692 190.377318 43.519692 62.292892 0 100.137957-43.647691 189.779989-43.647691 67.839519 0 139.732344 37.802399 190.889315 103.03927-167.678812 94.036667-140.543004 339.069598 28.885128 404.605134z"
                      fill="#0B0B0A"
                      p-id="1725"
                    ></path>
                  </svg>
                  <svg
                    v-if="u.provider == 'WXScan'"
                    class="w-full h-full"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                  >
                    <path fill="none" d="M0 0h24v24H0z" />
                    <path
                      d="M15.84 12.691l-.067.02a1.522 1.522 0 0 1-.414.062c-.61 0-.954-.412-.77-.921.136-.372.491-.686.925-.831.672-.245 1.142-.804 1.142-1.455 0-.877-.853-1.587-1.905-1.587s-1.904.71-1.904 1.587v4.868c0 1.17-.679 2.197-1.694 2.778a3.829 3.829 0 0 1-1.904.502c-1.984 0-3.598-1.471-3.598-3.28 0-.576.164-1.117.451-1.587.444-.73 1.184-1.287 2.07-1.541a1.55 1.55 0 0 1 .46-.073c.612 0 .958.414.773.924-.126.347-.466.645-.861.803a2.162 2.162 0 0 0-.139.052c-.628.26-1.061.798-1.061 1.422 0 .877.853 1.587 1.905 1.587s1.904-.71 1.904-1.587V9.566c0-1.17.679-2.197 1.694-2.778a3.829 3.829 0 0 1 1.904-.502c1.984 0 3.598 1.471 3.598 3.28 0 .576-.164 1.117-.451 1.587-.442.726-1.178 1.282-2.058 1.538zM2 12c0 5.523 4.477 10 10 10s10-4.477 10-10S17.523 2 12 2 2 6.477 2 12z"
                      fill="rgba(56,186,109,1)"
                    />
                  </svg>
                </div>
                <div class="relative">
                  <div class="items-center text-gray-900">
                    <p class="font-semibold hover:text-blue-500">
                      <a class="stretched-link">{{ u.name }} </a>
                    </p>
                    <span v-if="u.status == 'suspend'">
                      <div
                        class="inline-flex items-center align-middle justify-center font-medium border border-red-50 bg-red-50 text-red-600 rounded-sm px-1 text-xs mr-1"
                      >
                        已冻结
                      </div>
                    </span>
                  </div>
                  <div class="flex items-center text-gray-600 text-sm">
                    <span>{{ u.magicDomain }} </span>
                  </div>
                </div>
              </td>
              <td class="table-cell items-center md:w-1/4 lg:w-1/5">
                <div class="flex relative min-w-0">
                  <div class="truncate">
                    <span>{{ u.owner }} </span>
                  </div>
                </div>
              </td>
              <td class="hidden lg:table-cell items-center lg:w-1/5">
                <time :datetime="u.created" :title="u.created">{{
                  new Date(u.created)
                    .toLocaleDateString()
                    .replace("/", "年")
                    .replace("/", "月") + "日"
                }}</time>
              </td>
              <td class="hidden lg:table-cell items-center lg:w-1/5">
                <span>
                  <div class="inline-flex items-center cursor-default">
                    <span
                      class="inline-block w-2 h-2 rounded-full mr-2"
                      :class="{
                        'bg-green-500': u.currentlyConnected,
                        'bg-gray-300': !u.currentlyConnected,
                      }"
                    ></span>
                    <span
                      v-if="u.currentlyConnected"
                      class="text-sm text-gray-600 tooltip tooltip-top"
                      :data-tip="
                        '最近在线于' +
                        new Date(u.lastSeen)
                          .toLocaleString()
                          .replace('/', '年')
                          .replace('/', '月')
                          .replace(' ', '日 ')
                      "
                      >已连接</span
                    >
                    <span
                      v-else
                      class="text-sm text-gray-600 tooltip tooltip-top"
                      :data-tip="
                        new Date(u.lastSeen).getTime() ==
                        new Date('0001-01-01T00:00:00Z').getTime()
                          ? '未曾连接'
                          : '最近在线于' +
                            new Date(u.lastSeen)
                              .toLocaleString()
                              .replace('/', '年')
                              .replace('/', '月')
                              .replace(' ', '日 ')
                      "
                      >{{
                        new Date(u.lastSeen).getTime() ==
                        new Date("0001-01-01T00:00:00Z").getTime()
                          ? "未曾连接"
                          : new Date(u.lastSeen)
                              .toLocaleString()
                              .replace("/", "年")
                              .replace("/", "月")
                              .replace(" ", "日 ")
                      }}
                    </span>
                  </div>
                </span>
              </td>
              <td
                class="table-cell justify-end ml-auto md:ml-0 relative items-center w-8"
              >
                <div
                  v-if="(!tenantBtnShow && !tenantMenuShow) || selectTenant.id != u.id"
                  @click="openTenantMenu(u, $event)"
                  class="flex-none w-12 -mt-0.5 relative"
                >
                  <button
                    class="py-0.5 px-2 shadow-none rounded-md border border-gray-300/0 hover:border-gray-300/100 hover:bg-gray-100 hover:shadow-md hover:cursor-pointer active:border-gray-300/100 active:shadow focus:outline-none focus:ring transition-shadow duration-100 ease-in-out z-20"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="24"
                      height="24"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      class="text-gray-500"
                    >
                      <circle cx="12" cy="12" r="1"></circle>
                      <circle cx="19" cy="12" r="1"></circle>
                      <circle cx="5" cy="12" r="1"></circle>
                    </svg>
                  </button>
                </div>
                <!---->
                <div
                  v-if="(tenantBtnShow || tenantMenuShow) && selectTenant.id == u.id"
                  @click="openTenantMenu(u, $event)"
                  class="flex-none w-12 border button-outline bg-white shadow-md cursor-pointer focus:outline-none focus:ring -mt-0.5 relative py-0.5 px-2 rounded-md border-gray-300/100 hover:border-gray-300/100 hover:bg-gray-100 hover:shadow-md hover:cursor-pointer active:border-gray-300/100 transition-shadow duration-100 ease-in-out z-20"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="text-gray-500"
                  >
                    <circle cx="12" cy="12" r="1"></circle>
                    <circle cx="19" cy="12" r="1"></circle>
                    <circle cx="5" cy="12" r="1"></circle>
                  </svg>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </section>
  </main>

  <!-- 提示框显示 -->
  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>

  <!--设备配置菜单显示-->
  <Teleport to="body">
    <TenantMenu
      v-if="tenantMenuShow"
      :toleft="btnLeft"
      :totop="btnTop"
      :select-tenant="selectTenant"
      @close="closeTenantMenu"
      @showdialog-removetenant="showRemoveTenant"
    ></TenantMenu>
  </Teleport>

  <!-- 菜单弹出提示框显示 -->
  <Teleport to="body">
    <!-- 移除用户提示框显示 -->
    <RemoveTenant
      v-if="removeTenantShow"
      :select-tenant="selectTenant"
      @close="removeTenantShow = false"
      @confirm-remove="doRemoveTenant"
    >
    </RemoveTenant>
  </Teleport>
</template>

<style scoped>
.table tr.hover:hover th,
.table tr.hover:hover td,
.table tr.hover:nth-child(even):hover th,
.table tr.hover:nth-child(even):hover td {
  background-color: #faf9f8;
}

.table :where(thead, tfoot) :where(th, td) {
  background-color: #ffffff;
  color: #71706f;
  border-bottom-width: 1px;
}

.tooltip {
  --tooltip-color: #faf9f8;
  --tooltip-text-color: #3a3939;
  text-align: start;
  white-space: normal;
}

.tooltip:before {
  max-width: 16rem;
  font-size: small;
  font-weight: 300;
  border-radius: 0.375rem;
  box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  padding-left: 0.75rem;
  padding-right: 0.75rem;
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  border-width: 1px;
  border-color: #e1dfde;
}
</style>
