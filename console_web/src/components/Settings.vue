<script setup>
import { ref, computed, nextTick, onMounted, watch, watchEffect } from "vue";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import General from "./setpart/General.vue";
import Tags from "./setpart/Tags.vue";
import Keys from "./setpart/Keys.vue";
import DeviceManagement from "./setpart/DeviceManagement.vue";
import UserManagement from "./setpart/UserManagement.vue";

//路由及选择器页面控制
const setPartContent = {
  general: General,
  tags: Tags,
  "device-management": DeviceManagement,
  billing: General, //TODO
  "user-management": UserManagement,
  keys: Keys,
};
const route = useRoute();
const router = useRouter();
const currentSetPart = ref("");
function changeSetPart(event) {
  router.push("/settings/" + event.target.value);
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
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
              <polyline points="9 22 9 12 15 12 15 22"></polyline>
            </svg>
            <div class="ml-4">
              <h2 class="text-gray-500 font-medium">网络配置</h2>
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'general',
                  'text-gray-700': currentSetPart != 'general',
                }"
                to="/settings/general"
                >通用</router-link
              >
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'user-management',
                  'text-gray-700': currentSetPart != 'user-management',
                }"
                to="/settings/user-management"
                >用户管理</router-link
              >
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'device-management',
                  'text-gray-700': currentSetPart != 'device-management',
                }"
                to="/settings/device-management"
                >设备管理</router-link
              >
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'billing',
                  'text-gray-700': currentSetPart != 'billing',
                }"
                to="/settings/billing"
                >账单管理</router-link
              >
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'tags',
                  'text-gray-700': currentSetPart != 'tags',
                }"
                to="/settings/tags"
                >标签管理</router-link
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
              <h2 class="text-gray-500 font-medium">个人设置</h2>
              <router-link
                class="flex font-medium mt-4"
                :class="{
                  'text-blue-600': currentSetPart == 'keys',
                  'text-gray-700': currentSetPart != 'keys',
                }"
                to="/settings/keys"
                >密钥管理</router-link
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
            <optgroup label="网络配置">
              <option value="general">通用</option>
              <option value="user-management">用户管理</option>
              <option value="device-management">设备管理</option>
              <option value="billing">账单管理</option>
              <option value="tags">标签管理</option>
            </optgroup>
            <optgroup label="个人设置">
              <option value="keys">密钥管理</option>
            </optgroup>
          </select>
        </div>
      </div>

      <component :is="setPartContent[currentSetPart]"></component>
    </section>
  </main>
</template>

<style scoped></style>
