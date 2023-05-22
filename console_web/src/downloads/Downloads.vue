<script setup>
import { watch, ref, onMounted, computed } from "vue";
import { useRouter, useRoute } from "vue-router";

const route = useRoute();

const showMobileMenu = ref(false);
const download_details = ref({});

onMounted(() => {
  const jsonnode = document.getElementById("download-details");
  download_details.value = JSON.parse(jsonnode.innerText);
  window.addEventListener("resize", () => {
    if (window.innerWidth > 1024) {
      showMobileMenu.value = false;
    }
  });
});

const currentRoute = computed(() => {
  let curPath = route.path;
  return curPath;
});
</script>

<template>
  <div
    class="absolute inset-0 bg-[url(/img/grid.svg)] bg-center [mask-image:linear-gradient(180deg,white,rgba(255,255,255,0))]"
  ></div>
  <nav>
    <header>
      <div class="container flex items-center justify-between md:w-2/3">
        <div class="flex items-center">
          <a
            href="/"
            class="flex items-center mr-5 text-[#141414] transition-colors duration-200 z-10"
          >
            <svg class="p-2 block h-12 w-12">
              <use xlink:href="/img/icons.svg#logo"></use>
            </svg>
            <div class="text-3xl font-semibold truncate">蜃境</div>
          </a>
        </div>

        <button
          class="ml-auto lg:hidden z-10 text-gray-600 hover:text-gray-900"
          @click="showMobileMenu = true"
        >
          <svg class="p-2 block icon h-12 w-12">
            <use xlink:href="/img/icons.svg#menu"></use>
          </svg>
        </button>

        <div class="ml-auto hidden lg:block">
          <ul class="flex items-center">
            <li class="ml-1 md:ml-4 z-10">
              <a
                href="/admin"
                class="btn border-0 bg-blue-600 hover:bg-blue-700 text-white h-9 min-h-fit whitespace-nowrap text-sm font-medium sm:text-base"
              >
                <span>控制台</span>
              </a>
            </li>
          </ul>
        </div>
      </div>
    </header>

    <nav
      id="mobile-menu"
      :class="{
        hidden: !showMobileMenu,
      }"
      class="fixed top-0 right-0 -bottom-16 left-0 bg-gray-100 z-50 pt-2 container overflow-scroll overscroll-contain"
      style="margin-top: 0.15rem"
    >
      <div class="flex items-center justify-between mb-12">
        <a class="flex items-center" href="/">
          <svg class="p-2 block h-12 w-12">
            <use xlink:href="/img/icons.svg#logo"></use>
          </svg>
          <div class="text-3xl font-semibold truncate">蜃境</div>
        </a>
        <button
          class="-mr-2 lg:hidden text-gray-600 hover:text-gray-900"
          @click="showMobileMenu = false"
        >
          <svg class="p-2 block icon h-12 w-12">
            <use xlink:href="/img/icons.svg#x"></use>
          </svg>
        </button>
      </div>
      <div class="space-y-4 mb-8">
        <a
          href="/admin"
          class="flex w-full btn border-0 bg-blue-600 hover:bg-blue-700 text-white h-9 min-h-fit whitespace-nowrap"
        >
          <span>控制台</span>
        </a>
      </div>
    </nav>
  </nav>
  <header class="py-12">
    <div class="container text-center">
      <h1 class="f-articleTitle mb-3">下载蜃境客户端</h1>
      <h2 class="text-xl text-gray-600">安装应用并登录来开始使用吧</h2>
    </div>
  </header>
  <section class="mb-24 z-10">
    <div style="min-height: 75vh">
      <article
        class="py-6 px-2 sm:px-6 sm:rounded-lg shadow-2xl w-full sm:max-w-2xl mx-auto bg-white"
      >
        <nav class="text-center mb-7 overflow-y-hidden">
          <div class="flex overflow-y-auto overflow-x-scroll no-scrollbar">
            <router-link
              class="flex-grow flex-shrink-0 mx-1 block cursor-pointer select-none rounded pt-3 px-4 pb-2 font-normal text-center leading-normal align-middle focus:outline-none active:outline-none"
              :class="{
                'bg-blue-100 focus:bg-blue-100 active:bg-blue-100':
                  currentRoute == '/macOS',
                'focus:bg-blue-50 active:bg-blue-50': currentRoute != '/macOS',
              }"
              to="/macOS"
              ><svg class="inline-block mb-1" style="width: 32px; height: 20px">
                <use href="/img/platform-icons.svg#macos"></use>
              </svg>
              <div>macOS</div>
            </router-link>
            <router-link
              class="flex-grow flex-shrink-0 mx-1 block cursor-pointer select-none rounded pt-3 px-4 pb-2 font-normal text-center leading-normal align-middle focus:outline-none active:outline-none"
              :class="{
                'bg-blue-100 focus:bg-blue-100 active:bg-blue-100':
                  currentRoute == '/iOS',
                'focus:bg-blue-50 active:bg-blue-50': currentRoute != '/iOS',
              }"
              to="/iOS"
              ><svg class="inline-block mb-1" style="width: 32px; height: 20px">
                <use href="/img/platform-icons.svg#apple"></use>
              </svg>
              <div>iOS</div>
            </router-link>
            <router-link
              class="flex-grow flex-shrink-0 mx-1 block cursor-pointer select-none rounded pt-3 px-4 pb-2 font-normal text-center leading-normal align-middle focus:outline-none active:outline-none"
              :class="{
                'bg-blue-100 focus:bg-blue-100 active:bg-blue-100':
                  currentRoute == '/windows',
                'focus:bg-blue-50 active:bg-blue-50': currentRoute != '/windows',
              }"
              to="/windows"
              ><svg class="inline-block mb-1" style="width: 32px; height: 20px">
                <use href="/img/platform-icons.svg#windows"></use>
              </svg>
              <div>Windows</div>
            </router-link>
            <router-link
              class="flex-grow flex-shrink-0 mx-1 block cursor-pointer select-none rounded pt-3 px-4 pb-2 font-normal text-center leading-normal align-middle focus:outline-none active:outline-none"
              :class="{
                'bg-blue-100 focus:bg-blue-100 active:bg-blue-100':
                  currentRoute == '/linux',
                'focus:bg-blue-50 active:bg-blue-50': currentRoute != '/linux',
              }"
              to="/linux"
              ><svg class="inline-block mb-1" style="width: 32px; height: 20px">
                <use href="/img/platform-icons.svg#linux"></use>
              </svg>
              <div>Linux</div>
            </router-link>
            <router-link
              class="flex-grow flex-shrink-0 mx-1 block cursor-pointer select-none rounded pt-3 px-4 pb-2 font-normal text-center leading-normal align-middle focus:outline-none active:outline-none"
              :class="{
                'bg-blue-100 focus:bg-blue-100 active:bg-blue-100':
                  currentRoute == '/android',
                'focus:bg-blue-50 active:bg-blue-50': currentRoute != '/android',
              }"
              to="/android"
              ><svg class="inline-block mb-1" style="width: 32px; height: 20px">
                <use href="/img/platform-icons.svg#android"></use>
              </svg>
              <div>Android</div>
            </router-link>
          </div>
        </nav>
        <router-view
          class="px-3 md:px-0"
          :downloadDetails="download_details"
        ></router-view>
      </article>
    </div>
  </section>
</template>

<style scoped>
.icon {
  user-select: none;
  stroke: currentColor;
  stroke-width: 2;
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.f-articleTitle {
  font-size: 1.875rem;
  line-height: 2.25rem;
  font-weight: 500;
  line-height: 1.25;
  letter-spacing: -0.025em;
}
</style>
