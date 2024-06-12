import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'
import { useAuth } from '@/stores/auth.js'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
  ]
});

router.beforeEach(async to => {
  if (to.name !== 'login') {
    const auth = useAuth();
    if (!auth.jwt) {
      return '/login';
    }
  }
});

export default router
