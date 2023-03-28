<script setup>
import { watch, ref, onMounted, onBeforeUpdate, computed } from "vue";
import Toast from "../Toast.vue";
import SetTag from "./setDialog/SetTag.vue";

const devmode = ref(true);

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

const setTagShow = ref(false);
function showSetTag() {
  setTagShow.value = true;
}

const tagOwners = ref([]);

function createTagDone() {
  axios
    .get("/admin/api/acls/tags")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        tagOwners.value = response.data["data"]["tagOwners"];
      } else {
        toastMsg.value = "获取标签失败:" + response.data["status"].substring[6];
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      toastMsg.value = "获取标签失败:" + error;
      toastShow.value = true;
    });
}

onMounted(() => {
  axios
    .get("/admin/api/acls/tags")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        tagOwners.value = response.data["data"]["tagOwners"];
      } else {
        toastMsg.value = "获取标签失败:" + response.data["status"].substring[6];
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      toastMsg.value = "获取标签失败:" + error;
      toastShow.value = true;
    });
});

const wantRemoveTag = ref("");
const DeleteTagShow = ref(false);

function toRemoveTag(tag) {
  wantRemoveTag.value = tag;
  DeleteTagShow.value = true;
}

function doRemoveTag() {
  axios
    .delete("/admin/api/acls/tags/" + wantRemoveTag.value, {})
    .then(function (response) {
      if (response.data["status"] == "success") {
        var tmpTagOwners = [];
        for (var i in tagOwners.value) {
          if (tagOwners.value[i].tagName != "tag:" + response.data["data"]) {
            tmpTagOwners.push(tagOwners.value[i]);
          }
        }
        tagOwners.value = tmpTagOwners;
        DeleteTagShow.value = false;
      } else {
        toastMsg.value = "删除标签失败:" + response.data["status"].substring[6];
        toastShow.value = true;
      }
    })
    .catch(function (error) {
      // 处理错误情况
      toastMsg.value = "删除标签失败:" + error;
      toastShow.value = true;
    });
}
</script>

<template>
  <div class="flex-1">
    <div
      class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center"
    >
      <h1 class="mr-2" tabindex="-1">标签</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>查看和管理您的<strong>标签</strong>(<strong>Tag</strong>)</p>
      <p class="mt-2">
        使用标签可以很方便地设置服务应用类节点。它们通常不应归属于某个或某些用户，但代表部分用户可以访问到的提供业务服务的节点。
      </p>
      <p class="mt-2">
        一个标签的标签管理员意味着他可以将他所登录的节点打上对应的标签，从而由标签管理员负责某服务节点的部署配置。
      </p>
    </div>
    <div class="mt-10">
      <div class="flex justify-between items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">现有标签</h3>
          <p class="text-gray-600">
            设置的标签可在设备管理、授权密钥中分配，可在ACL中使用从而设置用户对服务的访问管理
          </p>
        </div>
        <button
          @click="showSetTag"
          class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3 font-normal"
        >
          创建标签…
        </button>
      </div>
      <div
        v-if="!tagOwners || tagOwners.length == 0"
        class="rounded-md border border-stone-200 mt-4 bg-stone-50 p-6"
      >
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">你还没有任何标签</div>
        </div>
      </div>

      <table
        v-if="tagOwners && tagOwners.length > 0"
        class="block border box-border rounded-lg mt-4 tb"
      >
        <thead
          class="block font-semibold tracking-wider text-left text-xs text-stone-500"
        >
          <tr class="flex border-b border-stone-200 pl-8 pr-4 lg:px-4">
            <th class="w-36 shrink-0 py-2">标签名称</th>
            <th class="flex-1 shrink-0 py-2 min-w-0">标签管理员</th>
            <th
              class="w-20 shrink-0 py-2 text-right text-red-400 cursor-pointer pointer-events-auto hover:text-red-600"
            >
              <span class="sr-only">删除标签</span>
            </th>
          </tr>
        </thead>
        <tbody class="block">
          <template v-for="(tag, i) in tagOwners">
            <tr
              :class="{ 'border-t': i > 0 }"
              class="group flex border-stone-200 cursor-pointer hover:bg-gray-50 pl-8 pr-4 lg:px-4 lg:cursor-auto border-b-0"
            >
              <td class="flex shrink-0 py-2 w-36">
                <pre
                  class="text-sm truncate leading-6 font-semibold"
                ><code>{{ tag.tagName.substring(4) }}</code></pre>
              </td>
              <td class="flex-1 shrink-0 py-2 min-w-0">
                <span v-for="(owner, j) in tag.owners">
                  <div
                    class="inline-flex items-center align-middle justify-center font-medium border border-stone-200 bg-stone-200 text-stone-600 rounded-sm px-1 text-xs mr-1"
                  >
                    {{ owner }}
                  </div>
                </span>
              </td>
              <td
                class="w-20 shrink-0 py-2 text-right text-red-400 cursor-pointer pointer-events-auto hover:text-red-600"
              >
                <button @click="toRemoveTag(tag.tagName.substring(4))" type="button">
                  删除…
                </button>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </div>
  <Teleport to="body">
    <!-- 创建Tag提示框显示 -->
    <SetTag
      v-if="setTagShow"
      @added-tag="createTagDone"
      @close="setTagShow = false"
    ></SetTag>
    <!-- 删除Tag提示框显示 -->
    <template v-if="DeleteTagShow">
      <div
        @click.self="DeleteTagShow = false"
        class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]"
        style="pointer-events: auto"
      >
        <div
          class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
          style="pointer-events: auto"
        >
          <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
            <div class="font-semibold text-lg truncate">注销</div>
          </header>
          <form @submit.prevent="doRemoveTag">
            <p class="text-gray-700 mb-4">
              删除此标签前请确保已经没有节点或ACL规则使用此标签
            </p>
            <footer class="flex mt-10 justify-end space-x-4">
              <button
                @click="DeleteTagShow = false"
                class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                type="button"
              >
                取消
              </button>
              <button
                class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit"
                type="submit"
              >
                删除标签
              </button>
            </footer>
          </form>
          <button
            @click="DeleteTagShow = false"
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
  </Teleport>

  <!-- 提示框显示 -->
  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>
</template>

<style scoped></style>
