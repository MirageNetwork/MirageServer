<script setup>
import { ref, watch, computed } from "vue";
import { useDisScroll } from "/src/utils.js";

const emit = defineEmits(["refresh-offers", "update-tcd"]);

useDisScroll();

const props = defineProps({
  currenttcd: String,
  tcdoffers: Array,
});
const wantedTCD = ref("");

function getTCDOffers() {
  wantedTCD.value = "";
  axios
    .get("/admin/api/tcd/offers")
    .then(function (response) {
      if (response.data["status"] == "success") {
        emit("refresh-offers", response.data["data"]["tcds"]);
      } else {
        console.log(response.data["status"]);
      }
    })
    .catch(function (error) {
      console.log(error);
    });
}

function setWantedTCD(newWantedTCD) {
  wantedTCD.value = newWantedTCD;
  console.log("new Wanted TCD: " + wantedTCD.value);
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
        <div class="font-semibold text-lg truncate">重命名 {{ currenttcd }}</div>
      </header>
      <form @submit.prevent="">
        <p v-if="tcdoffers.length == 0" class="mb-6">
          重命名你的蜃境可能会破坏你蜃境网络中已有的连接
        </p>
        <footer v-if="tcdoffers.length == 0" class="flex mt-10 justify-end space-x-4">
          <button
            @click="getTCDOffers"
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit w-full"
          >
            我明白，让我重命名蜃境
          </button>
        </footer>

        <div v-if="tcdoffers.length > 0" class="mt-8 mb-8">
          <p class="mb-5">
            选择一个生成的名称或者摇一摇获得新的一组。如果你选择了新的名称，你之前的名称会被释放给其他租户使用。
          </p>
          <div class="mb-5">
            <label v-for="(tcdOffer, i) in tcdoffers" class="flex mt-1">
              <input
                :checked="tcdOffer.tcd == wantedTCD"
                @change="setWantedTCD(tcdOffer.tcd)"
                type="radio"
                name="tcd-rad"
                class="radio radio-xs mt-1 mr-1"
                :value="tcdOffer.tcd"
              />{{ tcdOffer.tcd.split(".")[0]
              }}<span class="text-gray-400">{{
                tcdOffer.tcd.replace(tcdOffer.tcd.split(".")[0], "")
              }}</span>
            </label>
          </div>
          <button
            @click="getTCDOffers"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
            type="button"
          >
            <span>
              <svg
                width="1.25rem"
                height="1.25rem"
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                class="text-gray-500"
              >
                <path
                  fill="currentColor"
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M11.555 2.34556L11.5559 2.34511C11.694 2.27669 11.846 2.24109 12 2.24109C12.1541 2.24109 12.3061 2.27669 12.4442 2.34511L12.445 2.34556L19.759 6.00252L12 9.88198L4.24113 6.00252L11.555 2.34556ZM3.00005 7.61805V16.767C2.99875 16.9534 3.04954 17.1364 3.14672 17.2954C3.24374 17.4542 3.38313 17.5827 3.5492 17.6666L11 21.392V11.618L3.00005 7.61805ZM20.4428 17.6656L13 21.387V11.6181L21 7.61804V16.7695C20.9999 16.9555 20.948 17.1379 20.8499 17.296C20.7519 17.4541 20.6117 17.5817 20.445 17.6645L20.4428 17.6656ZM13.335 0.554497L12.89 1.45003L13.3373 0.555601L21.335 4.5545L21.3363 4.55513C21.8356 4.80352 22.2557 5.18614 22.5496 5.66006C22.8438 6.13439 22.9998 6.6819 23 7.24003V16.7706C22.9998 17.3287 22.8438 17.8757 22.5496 18.35C22.2557 18.824 21.8354 19.2067 21.336 19.4551L21.335 19.4556L13.3373 23.4545C12.9206 23.6629 12.461 23.7715 11.995 23.7715C11.529 23.7715 11.0693 23.6629 10.6525 23.4543L2.65283 19.4545L2.65007 19.4531C2.15077 19.2015 1.7317 18.8154 1.44015 18.3383C1.14927 17.8623 0.996878 17.3147 1.00005 16.7569L1.00005 7.2395C1.00034 6.68137 1.15633 6.13439 1.45047 5.66006C1.74436 5.18613 2.16454 4.80349 2.66381 4.55511L2.66505 4.5545L10.6628 0.555601L10.665 0.554497C11.0799 0.348361 11.5368 0.241089 12 0.241089C12.4633 0.241089 12.9202 0.348361 13.335 0.554497ZM12 7C12.8284 7 13.5 6.55228 13.5 6C13.5 5.44772 12.8284 5 12 5C11.1716 5 10.5 5.44772 10.5 6C10.5 6.55228 11.1716 7 12 7ZM5 14C5.55228 14 6 14.6716 6 15.5C6 16.3284 5.55228 17 5 17C4.44772 17 4 16.3284 4 15.5C4 14.6716 4.44772 14 5 14ZM10 13.5C10 12.6716 9.55228 12 9 12C8.44771 12 8 12.6716 8 13.5C8 14.3284 8.44771 15 9 15C9.55229 15 10 14.3284 10 13.5ZM15 12C15.5523 12 16 12.6716 16 13.5C16 14.3284 15.5523 15 15 15C14.4477 15 14 14.3284 14 13.5C14 12.6716 14.4477 12 15 12ZM16 17.5C16 16.6716 15.5523 16 15 16C14.4477 16 14 16.6716 14 17.5C14 18.3284 14.4477 19 15 19C15.5523 19 16 18.3284 16 17.5ZM19 14C19.5523 14 20 14.6716 20 15.5C20 16.3284 19.5523 17 19 17C18.4477 17 18 16.3284 18 15.5C18 14.6716 18.4477 14 19 14ZM20 11.5C20 10.6716 19.5523 10 19 10C18.4477 10 18 10.6716 18 11.5C18 12.3284 18.4477 13 19 13C19.5523 13 20 12.3284 20 11.5Z"
                ></path>
              </svg>
            </span>
            <span class="flex-1 ml-3">摇一摇</span>
          </button>
        </div>
        <footer v-if="tcdoffers.length > 0" class="flex mt-10 justify-end space-x-4">
          <button
            @click="$emit('close')"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            @click="$emit('update-tcd', wantedTCD)"
            :disabled="wantedTCD == ''"
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
          >
            保存
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
