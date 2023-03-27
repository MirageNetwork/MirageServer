<script setup>
import { ref, computed, nextTick, onMounted, watch, watchEffect } from "vue";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import General from "./setpart/General.vue";
import Authority from "./setpart/Authority.vue";
import RebindSA from "./setpart/RebindSA.vue";

//路由及选择器页面控制
const setPartContent = {
  general: General,
  authority: Authority,
  rebindSA: RebindSA,
};
const route = useRoute();
const router = useRouter();
const currentSetPart = ref("");
function changeSetPart(event) {
  router.push("/setting/" + event.target.value);
}

onBeforeRouteUpdate((to, from) => {
  currentSetPart.value = to.params.setpart;
});

onMounted(() => {
  currentSetPart.value = route.params.setpart;
});
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="md:flex md:mt-16">
      <div class="mb-10 md:mr-20 lg:mr-40">
        <div class="hidden md:block">
          <div class="flex flex-row mb-12">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="1.125em"
              height="1.125em"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-gray-500"
            >
              <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
              <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
              <line x1="6" y1="6" x2="6.01" y2="6"></line>
              <line x1="6" y1="18" x2="6.01" y2="18"></line>
            </svg>
            <div class="ml-4">
              <h2 class="text-gray-500 font-medium">系统配置</h2>
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'general',
                  'text-gray-700': currentSetPart != 'general',
                }"
                to="/setting/general"
                >基本配置</router-link
              >
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'authority',
                  'text-gray-700': currentSetPart != 'authority',
                }"
                to="/setting/authority"
                >第三方服务</router-link
              >
            </div>
          </div>
          <div class="flex flex-row mb-12">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-gray-500"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <div class="ml-4">
              <h2 class="text-gray-500 font-medium">其他配置</h2>
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'rebindSA',
                  'text-gray-700': currentSetPart != 'rebindSA',
                }"
                to="/setting/rebindSA"
                >换绑超管</router-link
              >
            </div>
          </div>
        </div>
        <div class="select-with-arrow md:hidden mb-4">
          <select
            v-model.lazy="currentSetPart"
            @change="changeSetPart"
            class="select select-bordered w-full text-lg"
          >
            <optgroup label="系统配置">
              <option value="general">基本配置</option>
              <option value="authority">第三方服务</option>
            </optgroup>
            <optgroup label="其他配置">
              <option value="rebindSA">换绑超管</option>
            </optgroup>
          </select>
        </div>
      </div>

      <component :is="setPartContent[currentSetPart]"></component>
    </section>
  </main>
</template>

<style scoped></style>
