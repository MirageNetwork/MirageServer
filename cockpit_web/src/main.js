import { createApp } from 'vue'
import { createMemoryHistory, createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import './style.css'
import App from './App.vue'
import Setting from './Settings.vue'
import RegAdmin from './RegAdmin.vue'
import Login from './Login.vue'
import VueClickAway from "vue3-click-away"


const routes = [
    { path: '/', redirect: 'setting' },
    { path: '/regAdmin', component: RegAdmin },
    { path: '/login', component: Login },
    { path: '/setting', redirect: '/setting/general' },
    { path: '/setting/:setpart', component: Setting },
]
const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

const app = createApp(App)
app.config.errorHandler = (err) => {
    /* 处理错误 */
    console.log(err)
}
app.use(router)
app.use(VueClickAway)
app.mount('#app-root')