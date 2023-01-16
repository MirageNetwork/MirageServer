import { createApp } from 'vue'
import { createMemoryHistory, createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import './style.css'
import App from './App.vue'
import Machines from './components/Machines.vue'
import Machine from './components/Machine.vue'
import Empty from './components/Empty.vue'

const routes = [
    { path: '/admin', component: Machines },
    { path: '/admin/machines', component: Machines },
    { path: '/admin/machines/:mip', component: Machine },
]


const router = createRouter({
    history: createWebHistory(),
    routes,
})

const app = createApp(App)
app.config.errorHandler = (err) => {
    /* 处理错误 */
}
app.use(router)
app.mount('#app')
