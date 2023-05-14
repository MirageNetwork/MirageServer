<script setup>
import { ref, computed, nextTick, onMounted, watch, watchEffect } from "vue";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";

import Tags from "./aclpart/Tags.vue";

//路由及选择器页面控制
const aclPartContent = {
  tags: Tags,
  option1: Tags,
  option2: Tags,
  option3: Tags,
};
const route = useRoute();
const router = useRouter();
const currentACLPart = ref("");
function changeACLPart(event) {
  router.push("/acls/" + event.target.value);
}

onBeforeRouteUpdate((to, from) => {
  currentACLPart.value = to.params.aclpart;
});

onMounted(() => {
  currentACLPart.value = route.params.aclpart;
});
</script>

<template>
  <main class="container mx-auto pb-20 md:pb-24">
    <section class="md:flex md:mt-16">
      <div class="mb-10 md:mr-20 lg:mr-40">
        <div class="hidden md:block">
          <div class="flex flex-row items-center mb-2">
            <svg
              viewBox="0 0 1024 1024"
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              fill="currentColor"
              stroke="currentColor"
              :stroke-width="currentACLPart == 'tags' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              :class="{
                'text-blue-600': currentACLPart == 'tags',
                'text-gray-700': currentACLPart != 'tags',
              }"
            >
              <path
                d="M947.2 422.4l-12.8-212.8c-3.2-64-56-116.8-120-120l-212.8-12.8c-36.8-1.6-72 11.2-97.6 36.8L113.6 502.4C64 553.6 64 633.6 113.6 683.2l225.6 225.6c49.6 49.6 131.2 49.6 180.8 0l388.8-388.8c27.2-25.6 41.6-60.8 38.4-97.6z m-81.6 52.8L475.2 865.6c-25.6 25.6-65.6 25.6-91.2 0L158.4 638.4c-25.6-25.6-25.6-65.6 0-91.2L548.8 158.4c12.8-12.8 30.4-19.2 49.6-19.2l214.4 12.8c32 1.6 57.6 27.2 60.8 60.8l12.8 212.8c-1.6 19.2-8 36.8-20.8 49.6z"
              ></path>
              <path
                d="M550.4 292.8c-49.6 49.6-49.6 131.2 0 180.8s131.2 49.6 180.8 0 49.6-131.2 0-180.8-131.2-49.6-180.8 0z m134.4 136c-25.6 25.6-65.6 25.6-91.2 0-25.6-25.6-25.6-65.6 0-91.2 25.6-25.6 65.6-25.6 91.2 0 25.6 25.6 25.6 65.6 0 91.2z"
              ></path>
            </svg>
            <!--
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentACLPart == 'tags' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              :class="{
                'text-blue-600': currentACLPart == 'tags',
                'text-gray-700': currentACLPart != 'tags',
              }"
            > 
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>-->
            <router-link
              class="flex font-medium ml-4"
              :class="{
                'text-blue-600': currentACLPart == 'tags',
                'text-gray-700': currentACLPart != 'tags',
              }"
              to="/acls/tags"
              >标签</router-link
            >
          </div>
          <div class="flex flex-row items-center mb-2">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentACLPart == 'option1' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              :class="{
                'text-blue-600': currentACLPart == 'option1',
                'text-gray-700': currentACLPart != 'option1',
              }"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <router-link
              class="flex font-medium ml-4"
              :class="{
                'text-blue-600': currentACLPart == 'option1',
                'text-gray-700': currentACLPart != 'option1',
              }"
              to="/acls/option1"
              >暂无内容</router-link
            >
          </div>
          <div class="flex flex-row items-center mb-2">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentACLPart == 'option2' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              :class="{
                'text-blue-600': currentACLPart == 'option2',
                'text-gray-700': currentACLPart != 'option2',
              }"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <router-link
              class="flex font-medium ml-4"
              :class="{
                'text-blue-600': currentACLPart == 'option2',
                'text-gray-700': currentACLPart != 'option2',
              }"
              to="/acls/option2"
              >暂无内容</router-link
            >
          </div>
          <div class="flex flex-row items-center mb-2">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              :stroke-width="currentACLPart == 'option3' ? '2.5' : '2'"
              stroke-linecap="round"
              stroke-linejoin="round"
              :class="{
                'text-blue-600': currentACLPart == 'option3',
                'text-gray-700': currentACLPart != 'option3',
              }"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <router-link
              class="flex font-medium ml-4"
              :class="{
                'text-blue-600': currentACLPart == 'option3',
                'text-gray-700': currentACLPart != 'option3',
              }"
              to="/acls/option3"
              >暂无内容</router-link
            >
          </div>
        </div>
        <div class="select-with-arrow md:hidden mb-4">
          <select
            v-model.lazy="currentACLPart"
            @change="changeACLPart"
            class="select select-bordered w-full text-lg"
          >
            <option value="tags">标签</option>
            <option value="option1">暂无内容</option>
            <option value="option2">暂无内容</option>
            <option value="option3">暂无内容</option>
          </select>
        </div>
      </div>

      <component :is="aclPartContent[currentACLPart]"></component>
    </section>
  </main>
</template>

<style scoped></style>
