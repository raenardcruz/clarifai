import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import DashboardView from './views/DashboardView.vue'
import RecordingsView from './views/RecordingsView.vue'
import RecordingShowView from './views/RecordingShowView.vue'
import SharedRecordingView from './views/SharedRecordingView.vue'
import SettingsView from './views/SettingsView.vue'
import UsersView from './views/UsersView.vue'

const routes = [
  { path: '/login', name: 'login', component: LoginView, meta: { requiresAuth: false } },
  { path: '/register', name: 'register', component: RegisterView, meta: { requiresAuth: false } },
  { path: '/share/:id', name: 'share', component: SharedRecordingView, meta: { requiresAuth: false } },
  { path: '/', name: 'dashboard', component: DashboardView, meta: { requiresAuth: true } },
  { path: '/recordings', name: 'recordings', component: RecordingsView, meta: { requiresAuth: true } },
  { path: '/recordings/:id', name: 'recording-show', component: RecordingShowView, meta: { requiresAuth: true } },
  { path: '/settings', name: 'settings', component: SettingsView, meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/users', name: 'users', component: UsersView, meta: { requiresAuth: true, requiresAdmin: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next('/')
  } else {
    next()
  }
})

export default router
