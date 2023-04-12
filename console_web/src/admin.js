import { createApp } from 'vue'
import { createMemoryHistory, createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import './style.css'
import App from './Console.vue'
import VueClickAway from "vue3-click-away"
import Machines from './components/Machines.vue'
import Machine from './components/Machine.vue'
import DNS from './components/DNS.vue'
import Settings from './components/Settings.vue'
import Users from './components/Users.vue'
import ACLs from './components/ACLs.vue'
import Navi from './components/Navi.vue'

const routes = [
    { path: '/', redirect: '/machines' },
    { path: '/machines', component: Machines },
    { path: '/machines/:mip', component: Machine },
    { path: '/dns', component: DNS },
    { path: '/settings', redirect: '/settings/general' },
    { path: '/settings/:setpart', component: Settings },
    { path: '/users', component: Users },
    { path: '/acls', redirect: '/acls/tags' },
    { path: '/acls/:aclpart', component: ACLs },
    { path: '/navi', component: Navi },
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
