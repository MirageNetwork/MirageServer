<script setup>
import { ref, computed, nextTick, onMounted, watch, watchEffect } from "vue";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import Toast from "./Toast.vue";

const devmode = ref("true")

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () { toastShow.value = false }, 5000)
  }
})
const DisableMagicDNSShow = ref(false)

const DNSCfg = ref({})

const MNetName = computed(() => {
  return DNSCfg.value["magicDNSDomains"][0]
})
const enOverride = computed(() => {
  return DNSCfg.value["resolvers"] && DNSCfg.value["resolvers"].length > 0
})
const enMagicDNS = computed(() => {
  return DNSCfg.value["magicDNS"]
})
const resolvers = computed(() => {
  if (enOverride.value) {
    return DNSCfg.value["resolvers"]
  }
  return DNSCfg.value["fallbackResolvers"]
})
const domains = computed(() => {
  return DNSCfg.value["domains"]
})
const domainResolvers = computed(() => {
  return DNSCfg.value["routes"]
})




const copyBtnText = ref("复制");

function copyMNetName() {
  navigator.clipboard.writeText(MNetName.value).then(function () {
    copyBtnText.value = "已复制!";
    setTimeout(() => {
      copyBtnText.value = "复制";
    }, 3000);
  });
}

onMounted(() => {
  axios
    .get("/admin/api/dns")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        DNSCfg.value = response.data["data"]
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
});

function switchMagicDNS(newStatus) {
  var reqData = DNSCfg.value
  switch (newStatus) {
    case "on":
    case "doOff":
      reqData["magicDNS"] = (newStatus == "on" ? true : false)
      axios
        .post("/admin/api/dns", reqData)
        .then(function (response) {
          if (response.data["status"] == "success") {
            DisableMagicDNSShow.value = false
            toastMsg.value = (newStatus == "on" ? "已启用幻域" : "已禁用幻域")
            toastShow.value = true
          } else {
            console.log(response.data["status"])
          }
        })
        .catch(function (error) {
          console.log(error)
        })
      break;
    case "off":
      DisableMagicDNSShow.value = true
      break;
  }
}
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="mb-24">
      <header class="mb-8">
        <div class="flex justify-between items-center">
          <div class="flex items-center">
            <h1 class="text-3xl font-semibold tracking-tight leading-tight mb-2" tabindex="-1">DNS</h1>
          </div>
        </div>
      </header>
      <section class="mb-16 max-w-2xl">
        <header class="mb-6">
          <h2 class="text-xl font-semibold tracking-tight mb-1">蜃境网域</h2>
          <p class="text-gray-600">这个名称用来注册DNS条目、分享设备给其他蜃境网域以及发放TLS证书</p>
        </header>
        <div class="max-w-sm">
          <div class="flex border border-stone-200 hover:border-stone-400 rounded-md relative overflow-hidden min-w-0">
            <input onclick="this.select()" class="outline-none py-2 px-3 w-full h-full font-mono text-sm text-ellipsis"
              readonly :value="MNetName" />
            <button @click="copyMNetName"
              class="flex justify-center py-2 pl-3 pr-4 rounded-md bg-white focus:outline-none font-sans text-blue-500 hover:text-blue-800 font-medium text-sm whitespace-nowrap">
              {{ copyBtnText }}
            </button>
          </div>
        </div>
        <button disabled="" @click="$emit('')"
          class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit mt-8">网域重命名</button>
      </section>
      <section class="mb-16 max-w-2xl">
        <header class="mb-6">
          <h2 class="text-xl font-semibold tracking-tight mb-1">幻域</h2>
          <p class="text-gray-600">自动为您蜃境网域中的设备注册域名，令您从而可以使用名称替代IP地址访问设备</p>
        </header>
        <button @click="switchMagicDNS('off')" v-if="enMagicDNS"
          class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit">停用幻域…</button>
        <button @click="switchMagicDNS('on')" v-if="!enMagicDNS"
          class="btn border-0 bg-blue-500 hover:bg-blue-900 text-white h-9 min-h-fit">启用幻域</button>
      </section>
      <section class="mb-16 max-w-2xl">
        <header class="mb-6">
          <h2 class="text-xl font-semibold tracking-tight mb-1">域名服务器</h2>
          <p class="text-gray-600">设置您网络中设备可用来解析DNS的域名服务器</p>
        </header>
        <div class="text-gray-900">
          <div v-if="enMagicDNS" class="py-3">
            <header class="flex mb-1">
              <h4 class="font-medium text-gray-600 mr-2 mb-1">{{ MNetName }}</h4>
              <div class="space-x-2 mb-1">
                <div
                  class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs relative -top-px">
                  <svg width="0.9em" height="0.9em" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"
                    class="mr-1">
                    <path
                      d="M6.5 12.5L8.05719 15.9428L11.5 17.5L8.05719 19.0572L6.5 22.5L4.94281 19.0572L1.5 17.5L4.94281 15.9428L6.5 12.5Z"
                      fill="currentColor"></path>
                    <path
                      d="M15.5 1L17.8358 6.16421L23 8.5L17.8358 10.8358L15.5 16L13.1642 10.8358L8 8.5L13.1642 6.16421L15.5 1Z"
                      fill="currentColor"></path>
                  </svg>
                  <span>幻域</span>
                </div>
              </div>
            </header>
            <div class="border border-gray-200 bg-white rounded-md divide-y overflow-hidden">
              <div class="flex justify-between select-none ">
                <div class="pl-4 flex flex-1 items-start">
                  <div class="tabular-nums pr-2 py-2 w-full">100.100.100.100</div>
                </div>
                <div class="p-2 pr-4 flex items-center justify-center"><svg xmlns="http://www.w3.org/2000/svg"
                    width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    stroke-linecap="round" stroke-linejoin="round" class="text-gray-400"
                    style="transform: translateX(0.125rem) scale(0.75);">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                  </svg></div>
              </div>
            </div>
          </div>
          <div data-rbd-droppable-id="search-domains" data-rbd-droppable-context-id="0">
            <div v-for="singleDomain in domains" class="DNSDomainGroup py-3 -mx-3 px-3 rounded-md"
              data-rbd-draggable-context-id="0" data-rbd-draggable-id="abc.com">
              <header class="flex items-center mb-1" tabindex="0" role="button"
                aria-describedby="rbd-hidden-text-0-hidden-text-0" data-rbd-drag-handle-draggable-id="abc.com"
                data-rbd-drag-handle-context-id="0" draggable="false">
                <div class="py-1 px-2 -ml-2"><svg xmlns="http://www.w3.org/2000/svg" width="1.2em" height="1.2em"
                    viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"
                    stroke-linejoin="round" class="relative -top-px text-gray-400">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="2" y1="12" x2="22" y2="12"></line>
                    <path
                      d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z">
                    </path>
                  </svg></div>
                <h4 class="font-medium mb-1 mr-2 text-gray-600">{{ singleDomain }}</h4>
                <div class="space-x-2 mb-1"><span data-state="closed">
                    <div
                      class="inline-flex items-center align-middle justify-center font-medium border border-gray-200 bg-gray-200 text-gray-600 rounded-sm px-1 text-xs">
                      <svg xmlns="http://www.w3.org/2000/svg" width="0.9em" height="0.9em" viewBox="0 0 24 24"
                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round" class="mr-1">
                        <polyline points="16 3 21 3 21 8"></polyline>
                        <line x1="4" y1="20" x2="21" y2="3"></line>
                        <polyline points="21 16 21 21 16 21"></polyline>
                        <line x1="15" y1="15" x2="21" y2="21"></line>
                        <line x1="4" y1="4" x2="9" y2="9"></line>
                      </svg>分离 DNS
                    </div>
                  </span></div>
              </header>
              <div v-for="oneRoute in domainResolvers[singleDomain]"
                class="border border-gray-200 bg-white rounded-md divide-y overflow-hidden">
                <div class="transition-shadow -mb-px flex justify-between select-none ">
                  <div class="pl-4 flex flex-1 items-start">
                    <div class="tabular-nums pr-2 py-2 w-full">{{ oneRoute }}</div>
                  </div>
                  <div class="pr-2 pt-1.5"><button type="button" id="radix-:r19:" aria-haspopup="menu"
                      aria-expanded="false" data-state="closed" class="py-0.5 px-2 shadow-none rounded-md border border-gray-300/0
          group-hover:border-gray-300/100 hover:border-gray-300/100 group-hover:bg-white hover:!bg-gray-0
          group-hover:shadow-md hover:shadow-md hover:cursor-pointer active:border-gray-300/100 active:shadow focus:outline-none focus:ring transition-shadow
          duration-100 ease-in-out z-50"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round" aria-label="Actions…" class="text-gray-500">
                        <circle cx="12" cy="12" r="1"></circle>
                        <circle cx="19" cy="12" r="1"></circle>
                        <circle cx="5" cy="12" r="1"></circle>
                      </svg></button></div>
                </div>
              </div>
            </div>
          </div>
          <div class="pt-3">
            <header class="flex items-center mb-2">
              <h4 class="font-medium text-gray-600 mr-2">全球域名服务器</h4>
              <div class="flex items-center space-x-2 ml-auto">
                <label class="flex items-center cursor-pointer text-sm font-medium select-none space-x-1 text-gray-500"
                  for="fallback">
                  <span class="tooltip" data-tip="启用时，连接的客户端会忽略本地DNS设置，并总使用这些全球域名服务器。
禁用时，客户端首选本地DNS设置，只在需要时使用这些全球服务器。">
                    <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="relative -top-0.5 inline-block">
                      <circle cx="12" cy="12" r="10"></circle>
                      <line x1="12" y1="16" x2="12" y2="12"></line>
                      <line x1="12" y1="8" x2="12.01" y2="8"></line>
                    </svg>
                  </span>
                  <span>覆盖本地 DNS</span>
                </label>
                <input :disabled="!resolvers || resolvers.length == 0" v-model="enOverride" id="fallback" ref="fallback"
                  type="checkbox" class="toggle toggle-xs">
              </div>
            </header>
            <div class="border border-gray-200 bg-white rounded-md divide-y overflow-hidden">
              <div v-if="!enOverride" class="-mb-px flex justify-between select-none ">
                <div class="pl-4 flex flex-1 items-start text-gray-400">
                  <div class="tabular-nums pr-2 py-2 w-full">本地 DNS 设置</div>
                </div>
                <div class="p-2 pr-4 flex items-center justify-center"><svg xmlns="http://www.w3.org/2000/svg"
                    width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    stroke-linecap="round" stroke-linejoin="round" class="text-gray-400"
                    style="transform: translateX(0.125rem) scale(0.75);">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                  </svg>
                </div>
              </div>
              <div v-for="ns in resolvers" class="transition-shadow -mb-px flex justify-between select-none ">
                <div class="pl-4 flex flex-1 items-start">
                  <div class="tabular-nums pr-2 py-2 w-full">{{ ns }}</div>
                </div>
                <div class="pr-2 pt-1.5"><button type="button" id="radix-:r26:" aria-haspopup="menu"
                    aria-expanded="false" data-state="closed" class="py-0.5 px-2 shadow-none rounded-md border border-gray-300/0
          group-hover:border-gray-300/100 hover:border-gray-300/100 group-hover:bg-white hover:!bg-gray-0
          group-hover:shadow-md hover:shadow-md hover:cursor-pointer active:border-gray-300/100 active:shadow focus:outline-none focus:ring transition-shadow
          duration-100 ease-in-out z-50"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                      viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                      stroke-linejoin="round" aria-label="Actions…" class="text-gray-500">
                      <circle cx="12" cy="12" r="1"></circle>
                      <circle cx="19" cy="12" r="1"></circle>
                      <circle cx="5" cy="12" r="1"></circle>
                    </svg></button>
                </div>
              </div>
            </div>
          </div>
        </div>
        <button
          class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit mt-8">
          添加域名服务器
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" class="ml-2">
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
      </section>

      <!--以下未实现-->
      <section v-if="!devmode" class="mb-16 max-w-2xl">
        <header class="mb-6">
          <h2 class="text-xl font-semibold tracking-tight mb-1">HTTPS 证书<div
              class="inline-flex items-center align-middle justify-center font-medium border border-yellow-50 bg-yellow-50 text-yellow-600 rounded-full px-2 py-1 leading-none relative text-sm ml-2 -top-px">
              Unavailable</div>
          </h2>
          <p class="text-gray-600">允许用户为他们的设备生辰HTTPS证书</p>
        </header>
        <button v-if="true"
          class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit">停用HTTPS证书…</button>
        <button v-if="false"
          class="btn border border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit">启用HTTPS证书…</button>
      </section>
    </section>
  </main>
  <!-- 提示框显示 -->
  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>

  <Teleport to="body">
    <!-- 停用幻域提示框显示 -->
    <template v-if="DisableMagicDNSShow">
      <div @click.self="DisableMagicDNSShow = false"
        class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]" style="pointer-events: auto;">
        <div
          class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
          style="pointer-events: auto;">
          <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
            <div class="font-semibold text-lg truncate">停用幻域？</div>
          </header>
          <form @submit.prevent="switchMagicDNS('doOff')">
            <p class="text-gray-700 mb-4">你网络中的用户将无法继续使用短名称在蜃境中访问设备</p>
            <footer class="flex mt-10 justify-end space-x-4">
              <button @click="DisableMagicDNSShow = false"
                class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                type="button">取消</button>
              <button class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit"
                type="submit">停用幻域</button>
            </footer>
          </form>
          <button @click="DisableMagicDNSShow = false"
            class="btn btn-sm btn-ghost absolute top-5 right-5 px-2 py-2 border-0 bg-base-0 focus:bg-base-200 hover:bg-base-200"
            type="button"><svg xmlns="http://www.w3.org/2000/svg" width="1.25em" height="1.25em" viewBox="0 0 24 24"
              fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg></button>
        </div>
      </div>
    </template>
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
  --togglehandleborder: 0 0 0 3px white inset, var(--handleoffsetcalculator) 0 0 3px white inset;
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
