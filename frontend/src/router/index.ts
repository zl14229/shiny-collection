import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { title: '仪表盘' },
        },
        {
          path: 'records',
          name: 'RecordList',
          component: () => import('@/views/RecordList.vue'),
          meta: { title: '狩猎记录' },
        },
        {
          path: 'records/create',
          name: 'RecordCreate',
          component: () => import('@/views/RecordCreate.vue'),
          meta: { title: '新增记录' },
        },
        {
          path: 'records/:id',
          name: 'RecordDetail',
          component: () => import('@/views/RecordDetail.vue'),
          meta: { title: '记录详情' },
        },
      ],
    },
  ],
})

export default router
