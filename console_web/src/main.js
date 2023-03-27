import { createApp } from 'vue'
import './style.css'
import Main from './Main.vue'

const app = createApp(Main)
app.config.errorHandler = (err) => {
    /* 处理错误 */
    console.log(err)
}
app.mount('body')
