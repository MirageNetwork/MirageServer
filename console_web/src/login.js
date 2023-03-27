import { createApp } from 'vue'
import './style.css'
import Login from './login/Login.vue'
import Register from './login/Register.vue'
import VueClickAway from "vue3-click-away"




const app = createApp(Login)
app.config.errorHandler = (err) => {
    /* 处理错误 */
    
}
app.use(VueClickAway)
app.mount('#app-root')
