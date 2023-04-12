<script setup>
import { ref, watch, computed, onMounted, nextTick } from "vue";
import { useDisScroll } from "/src/utils.js";
import Toast from "../Toast.vue";

const toastShow = ref(false);
const toastMsg = ref("");
watch(toastShow, () => {
  if (toastShow.value) {
    setTimeout(function () {
      toastShow.value = false;
    }, 5000);
  }
});

const emit = defineEmits(["add-done"]);

useDisScroll();

const props = defineProps({
  remotePubKey: String,
  naviRegionList: Array,
});
const orgRegionList = computed(() => {
  let result = [];
  if (!props.naviRegionList) {
    return result;
  }
  for (let i = 0; i < props.naviRegionList.length; i++) {
    if (props.naviRegionList[i]["Region"]["OrgID"] != 0) {
      result.push(props.naviRegionList[i]);
    }
  }
  return result;
});

const canAdd = computed(() => {
  console.log("willcreateregion ", willCreateRegion.value);
  console.log("selectRegion", selectRegion.value);
  return (
    (willCreateRegion.value
      ? newRegionCode.value != "" && newRegionName.value != ""
      : selectRegion.value["RegionCode"] && selectRegion.value["RegionCode"] != "") &&
    NewNaviNode.value["HostName"] &&
    NewNaviNode.value["HostName"] != "" &&
    (NewNaviNode.value["NoSTUN"] || NewNaviNode.value["STUNPort"] != "") &&
    (NewNaviNode.value["NoDERP"] || NewNaviNode.value["DERPPort"] != "") &&
    (extNavi.value ||
      (NewNaviNode.value["SSHAddr"] &&
        NewNaviNode.value["SSHAddr"] != "" &&
        (sshUsePassword.value
          ? NewNaviNode.value["SSHPwd"] && NewNaviNode.value["SSHPwd"] != ""
          : true)))
  );
});

onMounted(() => {});

// 新司南信息部分
const selectRegion = ref({});

const newRegionCode = ref("");
const newRegionName = ref("");
const willCreateRegion = ref(false);

const NewNaviNode = ref({});

// 远程主机参数部分
const extNavi = ref(false);

const selectDNSProvider = ref({});
const DNSChallenge = ref(true);
const DNSProviders = ref([
  { label: "Cloudflare", value: "cloudflare" },
  { label: "阿里云", value: "aliyun" },
  { label: "腾讯云", value: "qcloud" },
  { label: "NameSilo", value: "namesilo" },
]);

const sshUsePassword = ref(false);

const copyPubkeyBtnText = ref("复制");
function copyRemotePubkey(event) {
  navigator.clipboard.writeText(props.remotePubKey).then(function () {
    copyPubkeyBtnText.value = "已复制!";
    setTimeout(() => {
      copyPubkeyBtnText.value = "复制";
    }, 3000);
  });
}

const inDeploying = ref(false);

function doAddDerp() {
  inDeploying.value = true;
  console.log("Whether to create new region: ", willCreateRegion.value);
  console.log("New region select: ", selectRegion.value);

  NewNaviNode.value["RegionID"] = -1;
  var nodeRegionCode = newRegionCode.value;
  var nodeRegionName = newRegionName.value;
  if (willCreateRegion.value == false) {
    console.log("New region select ID is: ", selectRegion.value["RegionID"]);
    NewNaviNode.value["RegionID"] = selectRegion.value["RegionID"];
    nodeRegionCode = selectRegion.value["RegionCode"];
    nodeRegionName = selectRegion.value["RegionName"];
  }
  if (sshUsePassword.value == false) {
    NewNaviNode.value["SSHPwd"] = ""; //TEMP
  }

  NewNaviNode.value["DNSProvider"] = selectDNSProvider.value["value"];
  if (DNSChallenge.value == false) {
    NewNaviNode.value["DNSProvider"] = "";
    NewNaviNode.value["DNSID"] = "";
    NewNaviNode.value["DNSKey"] = "";
  }

  axios
    .post("/admin/api/derp/add", {
      RegionCode: nodeRegionCode,
      RegionName: nodeRegionName,
      NaviNode: NewNaviNode.value,
    })
    .then((res) => {
      console.log(res);
      if (res.data["status"] == "success") {
        inDeploying.value = false;
        emit("add-done", res.data["data"]);
      } else {
        inDeploying.value = false;
        toastShow.value = true;
        toastMsg.value = "添加失败:" + res.data["status"].substring(6);
      }
    })
    .catch((err) => {
      console.log(err);
      toastShow.value = true;
      toastMsg.value = "部署失败:" + err;
      inDeploying.value = false;
    });
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
        <div class="flex flex-row items-center">
          <div class="font-semibold text-lg truncate">添加新司南</div>
          <el-switch
            v-model="extNavi"
            size="large"
            class="ml-2 onetwoSwitch"
            inline-prompt
            active-text="登记"
            inactive-text="部署"
          />
        </div>
      </header>
      <form @submit.prevent="">
        <div class="flex flex-row w-full justify-between mb-2 space-x-2 items-center">
          <el-switch
            v-model="willCreateRegion"
            size="large"
            class="onetwoSwitch"
            inline-prompt
            active-text="新建区域"
            inactive-text="选择区域"
          />
          <label
            :class="{
              'swap-active': willCreateRegion,
            }"
            class="swap flex-inline w-2/3 text-md"
          >
            <el-select
              class="swap-off flex relative w-full z-20 justify-end"
              :teleported="false"
              :disabled="!orgRegionList || orgRegionList.length == 0"
              v-model="selectRegion"
              value-key="RegionID"
              filterable
              :placeholder="
                !orgRegionList || orgRegionList.length == 0
                  ? '不存在区域'
                  : '选择一个区域'
              "
            >
              <el-option
                v-for="nr in orgRegionList"
                :key="nr.Region.RegionID"
                :label="nr.Region.RegionName"
                :value="nr.Region"
              >
                <span style="float: left">{{ nr.Region.RegionName }}</span>
                <span
                  style="
                    float: right;
                    color: var(--el-text-color-secondary);
                    font-size: 13px;
                  "
                  >{{ nr.Region.RegionCode }}
                </span>
              </el-option>
            </el-select>
            <div class="swap-on flex flex-row relative w-full">
              <div class="flex w-2/5 justify-end mr-2 items-center">
                <p class="text-sm min-w-fit mr-1">代码</p>
                <div class="flex w-full">
                  <div class="relative w-full z-20">
                    <input
                      class="input w-full border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 text-sm min-h-fit"
                      type="text"
                      autocomplete="off"
                      autocorrect="off"
                      v-model="newRegionCode"
                    />
                  </div>
                </div>
              </div>
              <div class="flex w-3/5 justify-end items-center">
                <p class="text-sm min-w-fit mr-1">名称</p>
                <div class="flex w-full">
                  <div class="relative w-full z-20">
                    <input
                      class="input w-full border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 text-sm min-h-fit"
                      type="text"
                      autocomplete="off"
                      autocorrect="off"
                      v-model="newRegionName"
                    />
                  </div>
                </div>
              </div>
            </div>
          </label>
        </div>

        <div class="flex flex-row w-full mb-2 items-center">
          <p class="w-1/3 text-sm">域名</p>
          <div class="flex w-2/3">
            <div class="relative w-full z-10">
              <input
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                type="text"
                autocomplete="off"
                autocorrect="off"
                v-model="NewNaviNode.HostName"
              />
            </div>
          </div>
        </div>
        <div class="flex flex-row w-full mb-2 items-center">
          <p class="w-1/3 text-sm">IPv4 [可选]</p>
          <div class="flex w-2/3">
            <div class="relative w-full z-10">
              <input
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                type="text"
                autocomplete="off"
                autocorrect="off"
                v-model="NewNaviNode.IPv4"
              />
            </div>
          </div>
        </div>
        <div class="flex flex-row w-full mb-2 items-center">
          <p class="w-1/3 text-sm min-w-fit">IPv6 [可选]</p>
          <div class="flex w-2/3">
            <div class="relative w-full z-10">
              <input
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                type="text"
                autocomplete="off"
                autocorrect="off"
                v-model="NewNaviNode.IPv6"
              />
            </div>
          </div>
        </div>
        <div class="flex flex-row w-full mb-2 items-center">
          <el-switch
            v-model="NewNaviNode.NoSTUN"
            size="large"
            class="w-1/3 onoffSwitch"
            inline-prompt
            active-text="禁用探路"
            inactive-text="探路端口"
          />
          <div class="flex w-2/3">
            <div class="relative w-full z-10">
              <input
                :disabled="NewNaviNode.NoSTUN"
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                type="number"
                autocomplete="off"
                autocorrect="off"
                v-model="NewNaviNode.STUNPort"
              />
            </div>
          </div>
        </div>
        <div class="flex flex-row w-full mb-2 items-center">
          <el-switch
            v-model="NewNaviNode.NoDERP"
            size="large"
            class="w-1/3 onoffSwitch"
            inline-prompt
            active-text="禁用中继"
            inactive-text="中继端口"
          />
          <div class="flex w-2/3">
            <div class="relative w-full z-10">
              <input
                :disabled="NewNaviNode.NoDERP"
                class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                type="number"
                autocomplete="off"
                autocorrect="off"
                v-model="NewNaviNode.DERPPort"
              />
            </div>
          </div>
        </div>

        <div v-if="!extNavi">
          <div class="flex flex-row w-full items-center">
            <el-switch
              v-model="DNSChallenge"
              size="large"
              class="onetwoSwitch mr-2"
              inline-prompt
              active-text="新配DNS"
              inactive-text="已有DNS"
            />
            <el-divider content-position="left" class="flex items-center"
              >部署环境信息
            </el-divider>
          </div>
          <div v-if="DNSChallenge" class="flex flex-col w-full">
            <div class="flex flex-row w-full mb-2 items-center">
              <p class="w-1/3 text-sm">DNS 供应商</p>
              <el-select
                class="flex w-2/3 justify-end"
                :teleported="false"
                v-model="selectDNSProvider"
                value-key="value"
                filterable
                placeholder="选择一个供应商"
              >
                <el-option
                  v-for="dp in DNSProviders"
                  :key="dp.value"
                  :label="dp.label"
                  :value="dp"
                />
              </el-select>
            </div>
            <div class="flex flex-row w-full mb-2 items-center">
              <p class="w-1/3 text-sm">DNS API ID</p>
              <div class="flex w-2/3">
                <div class="relative w-full z-10">
                  <input
                    :disabled="
                      selectDNSProvider.value == 'cloudflare' ||
                      selectDNSProvider.value == 'namesilo'
                    "
                    class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                    type="text"
                    autocomplete="off"
                    autocorrect="off"
                    v-model="NewNaviNode.DNSID"
                  />
                </div>
              </div>
            </div>
            <div class="flex flex-row w-full mb-2 items-center">
              <p class="w-1/3 text-sm">DNS API Key</p>
              <div class="flex w-2/3">
                <div class="relative w-full z-10">
                  <input
                    class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                    type="text"
                    autocomplete="off"
                    autocorrect="off"
                    v-model="NewNaviNode.DNSKey"
                  />
                </div>
              </div>
            </div>
          </div>

          <div class="flex flex-row w-full mb-2 items-center">
            <p class="w-1/3 text-sm">SSH地址(带端口号)</p>
            <div class="flex w-2/3">
              <div class="relative w-full z-10">
                <input
                  class="input w-full z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                  type="text"
                  autocomplete="off"
                  autocorrect="off"
                  v-model="NewNaviNode.SSHAddr"
                />
              </div>
            </div>
          </div>
          <div class="flex flex-row w-full h-7 justify-start py-1 space-x-2 items-center">
            <el-switch
              v-model="sshUsePassword"
              size="large"
              class="onetwoSwitch z-30"
              inline-prompt
              active-text="使用口令"
              inactive-text="使用密钥"
            />
          </div>
          <label
            :class="{
              'swap-active': sshUsePassword,
            }"
            class="swap block w-full text-md -mt-7"
          >
            <div class="swap-on flex relative w-full z-10 justify-end">
              <input
                class="input w-2/3 z-30 border focus:outline-blue-500/60 hover:border disabled:hover:border-stone-200 disabled:border-stone-200 border-stone-200 hover:border-stone-400 rounded-md h-7 min-h-fit"
                type="password"
                autocomplete="off"
                autocorrect="off"
                v-model="NewNaviNode.SSHPwd"
              />
            </div>
            <div
              class="swap-off flex relative w-full -mt-7 justify-end pointer-events-none"
            >
              <div
                class="rounded-md border w-2/3 border-stone-200 gap-2 max-w-sm bg-stone-50 p-2"
              >
                <div class="flex flex-col justify-start">
                  <div class="w-full text-left font-bold max-w-xl text-gray-500">
                    添加到目标机器.ssh/authorized_keys
                  </div>
                  <div class="w-full text-left max-w-xl text-gray-500 break-all">
                    {{ remotePubKey }}
                  </div>
                  <button
                    @click="copyRemotePubkey($event)"
                    class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-7 min-h-fit"
                  >
                    {{ copyPubkeyBtnText }}
                  </button>
                </div>
              </div>
            </div>
          </label>
        </div>

        <footer class="flex mt-10 justify-end space-x-4">
          <button
            @click="$emit('close')"
            class="btn border border-stone-300 hover:border-stone-300 bg-stone-200 hover:bg-stone-300 text-black h-9 min-h-fit"
          >
            取消
          </button>
          <button
            :disabled="!canAdd || inDeploying"
            @click="doAddDerp"
            :class="{
              loading: inDeploying,
            }"
            class="btn border-0 bg-blue-500 hover:bg-blue-900 disabled:bg-blue-500/60 text-white disabled:text-white/60 h-9 min-h-fit"
          >
            添加
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

  <Teleport to=".toast-container">
    <Toast :show="toastShow" :msg="toastMsg" @close="toastShow = false"></Toast>
  </Teleport>
</template>

<style scoped>
input[type="number"]::-webkit-inner-spin-button,
input[type="number"]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type="number"] {
  -moz-appearance: textfield;
}
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

.onoffSwitch {
  --el-switch-on-color: #ff4949;
  --el-switch-off-color: #13ce66;
}
.offonSwitch {
  --el-switch-on-color: #13ce66;
  --el-switch-off-color: #ff4949;
}
.onetwoSwitch {
  --el-switch-on-color: #0082f6;
  --el-switch-off-color: #0082f6;
}

.swap {
  cursor: default;
}
</style>
