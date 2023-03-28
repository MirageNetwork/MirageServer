<script setup>
import { ref, watch, computed, onMounted } from "vue";
import { useDisScroll } from "/src/utils.js";

const emit = defineEmits(["change-role"]);

useDisScroll();

const props = defineProps({
  selectTenant: Object,
});

const newTenantCfg = ref({});
watch(
  () => props.selectTenant,
  (val) => {
    switch (currentTab) {
      case "basic":
        newTenantCfg.value["magicDomain"] = val["magicDomain"];
        newTenantCfg.value["owner"] = val["owner"];
        break;
      case "owner":
        newTenantCfg.value["magicDomain"] = val["magicDomain"];
        newTenantCfg.value["provider"] = val["provider"];
        newTenantCfg.value["name"] = val["name"];
        break;
      case "domain":
        newTenantCfg.value["owner"] = val["owner"];
        newTenantCfg.value["provider"] = val["provider"];
        newTenantCfg.value["name"] = val["name"];
        break;
    }
  }
);

const currentTab = ref("basic");

onMounted(() => {
  newTenantCfg.value["magicDomain"] = props.selectTenant["magicDomain"];
  newTenantCfg.value["owner"] = props.selectTenant["owner"];
  newTenantCfg.value["provider"] = props.selectTenant["provider"];
  newTenantCfg.value["name"] = props.selectTenant["name"];
});

function changeTab(tab) {
  currentTab.value = tab;
  newTenantCfg.value["magicDomain"] = props.selectTenant["magicDomain"];
  newTenantCfg.value["owner"] = props.selectTenant["owner"];
  newTenantCfg.value["provider"] = props.selectTenant["provider"];
  newTenantCfg.value["name"] = props.selectTenant["name"];
}
</script>

<template>
  <div
    @click.self="$emit('close')"
    class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
    style="pointer-events: auto"
  >
    <div
      class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
      tabindex="-1"
      style="pointer-events: auto"
    >
      <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
        <div class="font-semibold text-lg truncate">编辑租户配置</div>
      </header>
      <form @submit.prevent="">
        <div class="tabs">
          <a
            @click="changeTab('basic')"
            :class="{ 'tab-active': currentTab == 'basic' }"
            class="tab tab-bordered"
            >变更账号配置</a
          >
          <a
            @click="changeTab('owner')"
            :class="{ 'tab-active': currentTab == 'owner' }"
            class="tab tab-bordered"
            >变更所有者</a
          >
          <a
            @click="changeTab('domain')"
            :class="{ 'tab-active': currentTab == 'domain' }"
            class="tab tab-bordered"
            >变更蜃境域名</a
          >
        </div>
        <div v-if="currentTab == 'domain'">
          <p class="mb-2 mt-6">输入新的蜃境域名：</p>
          <div class="flex mb-2">
            <div class="relative w-full z-10">
              <input
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
                type="text"
                autocomplete="off"
                autocorrect="off"
                v-model="newTenantCfg.magicDomain"
              />
            </div>
          </div>
        </div>
        <div v-if="currentTab == 'owner'">
          <p class="mb-2 mt-6">输入新的所有者：</p>
          <div class="flex mb-2">
            <div class="relative w-full z-10">
              <input
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
                type="text"
                autocomplete="off"
                autocorrect="off"
                v-model="newTenantCfg.owner"
              />
            </div>
          </div>
        </div>
        <div v-if="currentTab == 'basic'">
          <p class="mb-2 mt-6">输入新的租户账号：</p>
          <div class="flex mb-2">
            <div class="relative w-full z-10">
              <input
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-9 min-h-fit"
                type="text"
                autocomplete="off"
                autocorrect="off"
                v-model="newTenantCfg.name"
              />
            </div>
          </div>
          <p class="mb-2 mt-6">选择身份商：</p>
          <div class="flex flex-row space-x-2">
            <label class="flex items-center space-x-2 mb-2">
              <input
                v-model="newTenantCfg.provider"
                type="radio"
                name="provider-rad"
                class="radio radio-xs mt-1 mr-1"
                value="Microsoft"
              /><svg
                class="w-6 h-6"
                viewBox="0 0 16 16"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path d="M0 0H7.57886V7.57886H0V0Z" fill="#F25022"></path>
                <path d="M0 8.42114H7.57886V16H0V8.42114Z" fill="#00A4EF"></path>
                <path d="M8.42114 0H16V7.57886H8.42114V0Z" fill="#7FBA00"></path>
                <path d="M8.42114 8.42114H16V16H8.42114V8.42114Z" fill="#FFB900"></path>
              </svg>
            </label>
            <label class="flex items-center space-x-2 mb-2">
              <input
                v-model="newTenantCfg.provider"
                type="radio"
                name="provider-rad"
                class="radio radio-xs mt-1 mr-1"
                value="Github" />
              <svg
                class="w-6 h-6"
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
                ></path></svg
            ></label>
            <label class="flex items-center space-x-2 mb-2">
              <input
                v-model="newTenantCfg.provider"
                type="radio"
                name="provider-rad"
                class="radio radio-xs mt-1 mr-1"
                value="Google"
              />
              <svg
                class="w-6 h-6"
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
            </label>
            <label class="flex items-center space-x-2 mb-2">
              <input
                v-model="newTenantCfg.provider"
                type="radio"
                name="provider-rad"
                class="radio radio-xs mt-1 mr-1"
                value="Apple"
              />
              <svg
                class="w-6 h-6"
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
            </label>
            <label class="flex items-center space-x-2 mb-2">
              <input
                v-model="newTenantCfg.provider"
                type="radio"
                name="provider-rad"
                class="radio radio-xs mt-1 mr-1"
                value="WXScan"
              />
              <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                <path fill="none" d="M0 0h24v24H0z" />
                <path
                  d="M15.84 12.691l-.067.02a1.522 1.522 0 0 1-.414.062c-.61 0-.954-.412-.77-.921.136-.372.491-.686.925-.831.672-.245 1.142-.804 1.142-1.455 0-.877-.853-1.587-1.905-1.587s-1.904.71-1.904 1.587v4.868c0 1.17-.679 2.197-1.694 2.778a3.829 3.829 0 0 1-1.904.502c-1.984 0-3.598-1.471-3.598-3.28 0-.576.164-1.117.451-1.587.444-.73 1.184-1.287 2.07-1.541a1.55 1.55 0 0 1 .46-.073c.612 0 .958.414.773.924-.126.347-.466.645-.861.803a2.162 2.162 0 0 0-.139.052c-.628.26-1.061.798-1.061 1.422 0 .877.853 1.587 1.905 1.587s1.904-.71 1.904-1.587V9.566c0-1.17.679-2.197 1.694-2.778a3.829 3.829 0 0 1 1.904-.502c1.984 0 3.598 1.471 3.598 3.28 0 .576-.164 1.117-.451 1.587-.442.726-1.178 1.282-2.058 1.538zM2 12c0 5.523 4.477 10 10 10s10-4.477 10-10S17.523 2 12 2 2 6.477 2 12z"
                  fill="rgba(56,186,109,1)"
                />
              </svg>
            </label>
          </div>
        </div>

        <footer class="flex mt-10 justify-end space-x-4">
          <button
            @click="$emit('close')"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            @click="$emit('update-tenant', newTenantCfg)"
            :disabled="
              selectTenant.name == newTenantCfg.name &&
              selectTenant.provider == newTenantCfg.provider &&
              selectTenant.owner == newTenantCfg.owner &&
              selectTenant.magicDomain == newTenantCfg.magicDomain
            "
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
          >
            更新租户配置
          </button>
        </footer>
      </form>
      <button
        @click="$emit('close')"
        class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
        type="button"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="1.25em"
          height="1.25em"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
  </div>
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

.radio {
  --chkbg: white;
  border-width: 2px;
  border-color: #d6d3d1;
}
.radio:checked {
  --chkbg: white;
  border-width: 5px;
  border-color: #3e5db3;
}
</style>
