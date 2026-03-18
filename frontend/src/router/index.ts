import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/Hello/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/blank',
      name: 'blank',
      component: () => import('@/views/Blank')
    },
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/skeleton',
      name: 'skeleton',
      component: () => import('@/views/Skeleton/SkeletonView.vue'),
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/AboutView.vue'),
    },
  ],
})

export default router
