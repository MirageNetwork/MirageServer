<script setup>
import { watch, onMounted } from 'vue'
const props = defineProps({
    identifyCode: {
        type: String,
        default: 'jida'
    },
    fontSizeMin: {
        type: Number,
        default: 25
    },
    fontSizeMax: {
        type: Number,
        default: 35
    },
    backgroundColorMin: {
        type: Number,
        default: 200
    },
    backgroundColorMax: {
        type: Number,
        default: 220
    },
    dotColorMin: {
        type: Number,
        default: 60
    },
    dotColorMax: {
        type: Number,
        default: 120
    },
    contentWidth: {
        type: Number,
        default: 80
    },
    contentHeight: {
        type: Number,
        default: 36
    }
},)



function randomNum(min, max) {
    return Math.floor(Math.random() * (max - min) + min)
}
function randomColor(min, max) {
    let r = randomNum(min, max)
    let g = randomNum(min, max)
    let b = randomNum(min, max)
    return 'rgb(' + r + ',' + g + ',' + b + ')'
}
function drawPic() {
    let canvas = document.getElementById('s-canvas')
    let ctx = canvas.getContext('2d')
    ctx.textBaseline = 'bottom'

    ctx.fillStyle = randomColor(props.backgroundColorMin, props.backgroundColorMax)
    ctx.fillRect(0, 0, props.contentWidth, props.contentHeight)

    for (let i = 0; i < props.identifyCode.length; i++) {
        drawText(ctx, props.identifyCode[i], i)
    }
    drawLine(ctx)
    drawDot(ctx)
}
function drawText(ctx, txt, i) {
    ctx.fillStyle = randomColor(50, 160)
    ctx.font = randomNum(props.fontSizeMin, props.fontSizeMax) + 'px SimHei'
    let x = (i + 1) * (props.contentWidth / (props.identifyCode.length + 1))
    let y = randomNum(props.fontSizeMax, props.contentHeight - 5)
    var deg = randomNum(-30, 30)

    ctx.translate(x, y)
    ctx.rotate(deg * Math.PI / 180)
    ctx.fillText(txt, 0, 0)

    ctx.rotate(-deg * Math.PI / 180)
    ctx.translate(-x, -y)
}
function drawLine(ctx) {
    for (let i = 0; i < 4; i++) {
        ctx.strokeStyle = randomColor(100, 200)
        ctx.beginPath()
        ctx.moveTo(randomNum(0, props.contentWidth), randomNum(0, props.contentHeight))
        ctx.lineTo(randomNum(0, props.contentWidth), randomNum(0, props.contentHeight))
        ctx.stroke()
    }
}
function drawDot(ctx) {
    for (let i = 0; i < 30; i++) {
        ctx.fillStyle = randomColor(0, 255)
        ctx.beginPath()
        ctx.arc(randomNum(0, props.contentWidth), randomNum(0, props.contentHeight), 1, 0, 2 * Math.PI)
        ctx.fill()
    }
}
watch(() => props.identifyCode, () => {
    drawPic()
}
)
onMounted(() => {
    drawPic()
})
</script>

<template>
    <div class="s-canvas -ml-20">
        <canvas id="s-canvas" class="rounded-r-md" :width="contentWidth" :height="contentHeight"></canvas>
    </div>
</template>

