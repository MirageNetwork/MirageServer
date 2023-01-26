<script setup>
import { computed, onMounted, ref, watch } from 'vue';

const emit = defineEmits(['focusSet', 'complete'])

//const pasteResult = ref([])
const props = defineProps({
    switchoff: Boolean,
    setFocus: Boolean
})
watch(() => props.setFocus, (newV) => {
    if (newV) {
        emit('focusSet')
        firstInput.value?.focus()
    }
})
const firstInput = ref(null)
// const vcin0 = ref(null)
// const vcin1 = ref(null)
// const vcin2 = ref(null)
// const vcin3 = ref(null)
// const vcin4 = ref(null)
// const vcin5 = ref(null)

const input = ref(Array(6))
/*
const input = computed({
    get() {
        if (props.code && Array.isArray(props.code) && props.code.length === 6) {
            return props.code;
        } else if (/^\d{6}$/.test(props.code.toString())) {
            return props.code.toString().split('');
        } else if (pasteResult.value.length === 6) {
            return pasteResult.value;
        } else {
            return new Array(6);
        }
    },
    set(a, b) {
        console.log(a + "*********" + b)
    }
})
*/

function inputEvent(e) {
    var index = e.target.dataset.index * 1;
    var el = e.target;
    el.value = el.value.replace(/Digit|Numpad/i, '').slice(0, 1); //.replace(/1/g, '')
    input.value[index] = el.value
    if (input.value.join('').length === 6) {
        document.activeElement.blur();
        emit('complete', input.value);
    }
}
function keydown(e) {
    var index = e.target.dataset.index * 1;
    var el = e.target;
    if (e.key === 'Backspace') {
        if (input.value[index].length > 0) {
            input.value[index] = ''
            emit('complete', []);
        } else {
            if (el.previousElementSibling) {
                el.previousElementSibling.focus()
                input.value[index - 1] = ''
                emit('complete', []);
            }
        }
    } else if (e.key === 'Delete') {
        if (input.value[index].length > 0) {
            input.value[index] = ''
            emit('complete', []);
        } else {
            if (el.nextElementSibling) {
                input.value[1] = ''
                emit('complete', []);
            }
        }
        if (el.nextElementSibling) {
            el.nextElementSibling.focus()
        }
    } else if (e.key === 'Home') {
        el.parentElement.children[0] && el.parentElement.children[0].focus()
    } else if (e.key === 'End') {
        el.parentElement.children[input.value.length - 1] && el.parentElement.children[input.value.length - 1].focus()
    } else if (e.key === 'ArrowLeft') {
        if (el.previousElementSibling) {
            el.previousElementSibling.focus()
        }
    } else if (e.key === 'ArrowRight') {
        if (el.nextElementSibling) {
            el.nextElementSibling.focus()
        }
        //    } else if (e.key === 'ArrowUp') {
        //        if (input.value[index] * 1 < 9) {
        //            input.value[index] = (input.value[index] * 1 + 1).toString();
        //        }
        //    } else if (e.key === 'ArrowDown') {
        //        if (input.value[index] * 1 > 0) {
        //            input.value[index] = (input.value[index] * 1 - 1).toString();
        //        }
    }
}
function keyup(e) {
    var index = e.target.dataset.index * 1;
    var el = e.target;
    // 解决输入e的问题
    el.value = el.value.replace(/Digit|Numpad/i, '').slice(0, 1); //.replace(/1/g, '')
    if (/Digit|Numpad/i.test(e.code)) {
        // 必须在这里赋值，否则输入框会是空值
        input.value[index] = e.code.replace(/Digit|Numpad/i, '');
        el.nextElementSibling && el.nextElementSibling.focus();
        if (input.value.join('').length === 6) {
            document.activeElement.blur();
            emit('complete', input.value);
        }
    } else {
        if (input.value[index] === '') {
            input.value[index] = '';
            emit('complete', []);
        }
    }
}
/*
function mousewheel(e) {
    var index = e.target.dataset.index;
    if (e.wheelDelta > 0) {
        if (input.value[index] * 1 < 9) {
            input.value[index](input.value[index] * 1 + 1).toString();
        }
    } else if (e.wheelDelta < 0) {
        if (input.value[index] * 1 > 0) {
            input.value[index](input.value[index] * 1 - 1).toString();
        }
    } else if (e.key === 'Enter') {
        if (input.value.join('').length === 6) {
            document.activeElement.blur();
            emit('complete', input.value);
        }
    }
}
*/
function paste(e) {
    // 当进行粘贴时
    e.clipboardData.items[0].getAsString(str => {
        if (str.toString().length === 6 && /^\d{6}$/.test(str.toString())) {
            input.value = str.split('');
            document.activeElement.blur();
            emit('complete', input.value);
        } else {
            // 如果粘贴内容不合规，清除所有内容
            input.value[0] = new Array(6)
            emit('complete', []);
        }
    })
}
onMounted(() => {
    // 等待dom渲染完成，在执行focus,否则无法获取到焦点
    //    firstinput.value.focus()
})
</script>
<template>
    <div class="flex space-x-3 " @keydown="keydown" @keyup="keyup" @paste="paste" @input="inputEvent">
        <input :disabled="switchoff" ref="firstInput"
            class="ml-2 border-1 border-stone-200 bg-stone-100 rounded-md shadow-inner h-9 w-7 text-2xl text-center"
            max="9" min="0" maxlength="1" data-index="0" v-model.trim.number="input[0]" type="number" />
        <input :disabled="switchoff"
            class="border-1 border-stone-200 bg-stone-100 rounded-md shadow-inner h-9 w-7 text-2xl text-center" max="9"
            min="0" maxlength="1" data-index="1" v-model.trim.number="input[1]" type="number" />
        <input :disabled="switchoff"
            class="border-1 border-stone-200 bg-stone-100 rounded-md shadow-inner h-9 w-7 text-2xl text-center" max="9"
            min="0" maxlength="1" data-index="2" v-model.trim.number="input[2]" type="number" />
        <input :disabled="switchoff"
            class="border-1 border-stone-200 bg-stone-100 rounded-md shadow-inner h-9 w-7 text-2xl text-center" max="9"
            min="0" maxlength="1" data-index="3" v-model.trim.number="input[3]" type="number" />
        <input :disabled="switchoff"
            class="border-1 border-stone-200 bg-stone-100 rounded-md shadow-inner h-9 w-7 text-2xl text-center" max="9"
            min="0" maxlength="1" data-index="4" v-model.trim.number="input[4]" type="number" />
        <input :disabled="switchoff"
            class="border-1 border-stone-200 bg-stone-100 rounded-md shadow-inner h-9 w-7 text-2xl text-center" max="9"
            min="0" maxlength="1" data-index="5" v-model.trim.number="input[5]" type="number" />
    </div>
</template>
<style>
input[type='number']::-webkit-outer-spin-button,
input[type='number']::-webkit-inner-spin-button {
    -webkit-appearance: none !important;
}

input[type='number'] {
    caret-color: transparent;
    -moz-appearance: textfield;
}
</style>
