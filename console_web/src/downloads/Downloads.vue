<script setup>
import { watch, ref, onMounted, computed } from "vue";
const showMobileMenu = ref(false);

onMounted(() => {
  window.addEventListener("resize", () => {
    if (window.innerWidth > 1024) {
      showMobileMenu.value = false;
    }
  });
});
</script>

<template>
  <div class="relative flex min-h-screen flex-col bg-gray-200">
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
  </div>
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
</style>
