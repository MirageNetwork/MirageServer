import { createApp } from 'vue'
import { createMemoryHistory, createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import './style.css'
import App from './Console.vue'
import VueClickAway from "vue3-click-away"
import Machines from './components/Machines.vue'
import Machine from './components/Machine.vue'
import Settings from './components/Settings.vue'


const routes = [
    { path: '/', redirect: '/machines' },
    { path: '/machines', component: Machines },
    { path: '/machines/:mip', component: Machine },
    { path: '/settings', redirect: '/settings/general' },
    { path: '/settings/:setpart', component: Settings },
]
const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

const app = createApp(App)
app.config.errorHandler = (err) => {
    /* 处理错误 */
}
app.use(VueClickAway)
app.use(router)
app.mount('#app')
