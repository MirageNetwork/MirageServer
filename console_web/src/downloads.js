import { createApp } from 'vue'
import { createMemoryHistory, createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import './style.css'
import Downloads from './downloads/Downloads.vue'
import Windows from './downloads/Windows.vue'
import IOS from './downloads/IOS.vue'
import Linux from './downloads/Linux.vue'
import Mac from './downloads/Mac.vue'
import Android from './downloads/Android.vue'

const routes = [
    { path: '/', redirect: '/win' },
    { path: '/win', component: Windows },
    { path: '/ios', component: IOS },
    { path: '/linux', component: Linux },
    { path: '/mac', component: Mac },
    { path: '/android', component: Android },
]
const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

const app = createApp(Downloads)
app.config.errorHandler = (err) => {
    /* 处理错误 */
    
}
app.use(router)
app.mount('body')
