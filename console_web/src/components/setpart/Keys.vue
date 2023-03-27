<script setup>
import { watch, ref, onMounted, onBeforeUpdate, computed } from 'vue';
import GenAuthKey from './setDialog/GenAuthKey.vue';

const devmode = ref(true)

const genAuthKeyShow = ref(false);
function showGenAuthKey() {
  genAuthKeyShow.value = true;
}

const tagOwners = ref([])
const authKeys = ref([])

function doAddAuthkey() {
  axios
    .get("/admin/api/keys")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        authKeys.value = response.data["data"]["authKeys"]
      }
    })
    .catch(function (error) {
      // 处理错误情况
      console.log(error);
    })
}

onMounted(() => {
  axios
    .get("/admin/api/keys")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        authKeys.value = response.data["data"]["authKeys"]
      }
    })
    .catch(function (error) {
      // 处理错误情况
      console.log(error);
    })
  axios
    .get("/admin/api/acls/tags")
    .then(function (response) {
      // 处理成功情况
      if (response.data["status"] == "success") {
        tagOwners.value = response.data["data"]["tagOwners"]
      }
    })
    .catch(function (error) {
      // 处理错误情况
      console.log(error);
    })
})

const wantRevokeAuthKeyID = ref("")
const RevokeAuthKeyShow = ref(false)

function toRevokeAuthKey(keyID) {
  wantRevokeAuthKeyID.value = keyID
  RevokeAuthKeyShow.value = true
}

function doRevokeAuthKey() {
  axios
    .delete("/admin/api/keys/" + wantRevokeAuthKeyID.value, {})
    .then(function (response) {
      if (response.data["status"] == "success") {
        var tmpAuthKeys = []
        for (var i in authKeys.value) {
          if (authKeys.value[i].id != response.data["data"]) {
            tmpAuthKeys.push(authKeys.value[i])
          }
        }
        authKeys.value = tmpAuthKeys
        RevokeAuthKeyShow.value = false
      } else {
        console.log(response.data["status"])
      }
    })
    .catch(function (error) {
      console.log(error)
    })
}

function isInvalidTag(tag)  {
  for (var i in tagOwners.value) {
    if (tagOwners.value[i].tagName == tag) {
      return false
    }
  }
  return true
}
</script>

<template>
  <div class="flex-1">
    <div class="text-3xl font-semibold tracking-tight leading-tight mb-2 flex items-center">
      <h1 class="mr-2" tabindex="-1">密钥管理</h1>
    </div>
    <div class="text-gray-600 mt-3 mb-10">
      <p>查看和管理您的<strong>授权密钥</strong>和<strong>API密钥</strong></p>
      <p class="mt-2">您个人设备的私钥<strong>不</strong>在此管理，它们永远是<strong>私有</strong>并存在于您的设备上，<strong>不</strong>向蜃境披露</p>
    </div>
    <div class="mt-10">
      <div class="flex justify-between items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">授权密钥</h3>
          <p class="text-gray-600">用于无需交互式登录即可实现设备鉴别接入 </p>
        </div>
        <button @click="showGenAuthKey"
          class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3 font-normal">
          生成授权密钥…</button>
      </div>
      <div v-if="!authKeys || authKeys.length == 0" class="rounded-md border border-stone-200 mt-4 bg-stone-50 p-6">
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">你还没有任何密钥</div>
        </div>
      </div>

      <table v-if="authKeys && authKeys.length > 0" class="block border box-border rounded-lg mt-4 tb">
        <thead class="block font-semibold tracking-wider text-left text-xs text-stone-500">
          <tr class="flex border-b border-stone-200 pl-8 pr-4">
            <th class="w-36 shrink-0 py-2">ID</th>
            <th class="hidden shrink-0 py-2 lg:block w-40">创建日期</th>
            <th class="hidden shrink-0 py-2 lg:block w-40">失效日期</th>
            <th class="flex-1 shrink-0 py-2 min-w-0">类型</th>
            <th class="w-20 shrink-0 py-2 text-right text-red-400 cursor-pointer pointer-events-auto hover:text-red-600">
              <span class="sr-only">注销密钥</span>
            </th>
          </tr>
        </thead>
        <tbody class="block">
          <template v-for="authKey, id in authKeys">
            <tr @click="authKey.expand = authKey.expand ? false : true"
              class="group flex border-stone-200 cursor-pointer hover:bg-gray-50 pl-8 pr-4 border-b-0" :class="{
                'border-t': id > 0,
                'lg:cursor-auto': !authKey.authkey.tags || authKey.authkey.tags.length == 0,
              }">
              <td class="flex shrink-0 py-2 w-36">
                <div class="w-8 -ml-8" :class="{
                  'lg:hidden': !authKey.authkey.tags || authKey.authkey.tags.length == 0,
                }"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"
                    class="text-gray-400 w-4 mx-2 group-hover:text-gray-500">
                    <polyline :points="authKey.expand ? '6 9 12 15 18 9' : '9 18 15 12 9 6'"></polyline>
                  </svg></div>
                <pre class="text-sm truncate leading-6 font-semibold"><code>{{ authKey.id }}</code></pre>
              </td>
              <td class="hidden shrink-0 py-2 lg:block w-40"><span data-state="closed"><span class="cursor-default">
                    {{ authKey.created.split(' ')[0] }}</span></span></td>
              <td class="hidden shrink-0 py-2 lg:block w-40"><span data-state="closed"><span class="cursor-default">
                    {{ authKey.expiry.split(' ')[0] }}</span></span></td>
              <td class="flex-1 shrink-0 py-2 min-w-0">{{
                (authKey.authkey.reusable == true ? "可重用" : "一次性") + (authKey.authkey.ephemeral == true ? ",自熄" : "") }}
              </td>
              <td
                class="w-20 shrink-0 py-2 text-right text-red-400 cursor-pointer pointer-events-auto hover:text-red-600">
                <button @click="toRevokeAuthKey(authKey.id)" type="button">注销…</button>
              </td>
            </tr>
            <tr v-if="authKey.expand" class="border-b-0 flex border-stone-200">
              <td class="shrink-0 p-0">
                <div class="flex-col px-8 pt-0.5 pb-2" :class="{
                  'lg:hidden': !authKey.authkey.tags || authKey.authkey.tags.length == 0,
                }">
                  <div class="flex items-center lg:hidden">
                    <p class="w-20 text-sm text-gray-500">创建日期</p>
                    <p class="text-sm"><span data-state="closed"><span class="cursor-default">{{
                      authKey.created.split(' ')[0]
                    }}</span></span>
                    </p>
                  </div>
                  <div class="flex items-center lg:hidden">
                    <p class="w-20 text-sm text-gray-500">失效日期</p>
                    <p class="text-sm"><span data-state="closed"><span class="cursor-default">{{
                      authKey.expiry.split(' ')[0]
                    }}</span></span>
                    </p>
                  </div>
                  <div class="flex items-center">
                    <p class="w-20 text-sm text-gray-500">标签</p>
                    <div v-for="tag, i in authKey.authkey.tags" class="flex flex-wrap">
                      <span><span>
                          <div
                            class="flex items-center align-middle justify-center font-medium border rounded-full px-2 py-1 leading-none text-xs mr-1"
                            :class="{
                              'border-gray-200 bg-gray-200 text-gray-600':isInvalidTag(tag),
                              'border-gray-300 bg-white':!isInvalidTag(tag),
                            }"
                            >
                            <svg v-if="isInvalidTag(tag)" xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none"
                              stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"
                              class="mr-1 text-gray-500">
                              <circle cx="12" cy="12" r="10"></circle>
                              <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line>
                            </svg>
                            <span class="text-gray-500">{{ tag.substring(4) }}</span>
                          </div>
                        </span></span>
                    </div>
                  </div>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>

      <!--以下API 密钥部分-->
      <div class="flex justify-between items-center mt-16">
        <div>
          <h3 class="text-xl font-semibold tracking-tight">API 密钥</h3>
          <p class="text-gray-600">API 密钥用于访问蜃境API.</p>
        </div>
        <button :disabled="devmode"
          class="btn border border-stone-300 hover:border-stone-300 disabled:border-stone-300 bg-base-200 hover:bg-base-300 disabled:bg-base-200/60 text-black disabled:text-black/30 h-9 min-h-fit ml-3 font-normal">
          生成 API 密钥…</button>
      </div>
      <div class="rounded-md border border-stone-200 mt-4 bg-stone-50 p-6">
        <div class="flex justify-center">
          <div class="w-full text-center max-w-xl text-gray-500">你还没有任何密钥</div>
        </div>
      </div>
    </div>
  </div>
  <Teleport to="body">
    <!-- 生成授权密钥提示框显示 -->
    <GenAuthKey v-if="genAuthKeyShow" :tag-owners="tagOwners" @added-authkey="doAddAuthkey"
      @close="genAuthKeyShow = false"></GenAuthKey>
    <!-- 注销授权密钥提示框显示 -->
    <template v-if="RevokeAuthKeyShow">
      <div @click.self="RevokeAuthKeyShow = false"
        class="fixed overflow-y-auto inset-0 py-8 z-30 bg-gray-900 bg-opacity-[0.07]" style="pointer-events: auto;">
        <div
          class="bg-white rounded-lg relative p-4 md:p-6 text-gray-700 max-w-lg min-w-[19rem] my-8 mx-auto w-[97%] shadow-2xl"
          style="pointer-events: auto;">
          <header class="flex items-center justify-between space-x-4 mb-5 mr-8">
            <div class="font-semibold text-lg truncate">注销</div>
          </header>
          <form @submit.prevent="doRevokeAuthKey">
            <p class="text-gray-700 mb-4">注销此密钥<strong>并不会</strong>注销已使用此密钥进行授权的设备，但将阻止之后继续使用此密钥注册新设备。</p>
            <footer class="flex mt-10 justify-end space-x-4">
              <button @click="RevokeAuthKeyShow = false"
                class="btn border border-base-300 hover:border-base-300 bg-base-200 hover:bg-base-300 text-black h-9 min-h-fit"
                type="button">取消</button>
              <button class="btn border-0 bg-red-600 hover:bg-red-700 text-white h-9 min-h-fit"
                type="submit">注销密钥</button>
            </footer>
          </form>
          <button @click="RevokeAuthKeyShow = false"
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

<style scoped></style>