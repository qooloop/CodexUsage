// src/router/index.ts
// 4 个视图：总览 / 统计详情 / 设置 / 关于
// 全部走 hash 模式（兼容 wails file://）

import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/overview' },
  { path: '/overview', name: 'overview', component: () => import('@/views/Dashboard.vue'), meta: { title: '总览' } },
  { path: '/stats', name: 'stats', component: () => import('@/views/Stats.vue'), meta: { title: '统计详情' } },
  { path: '/settings', name: 'settings', component: () => import('@/views/Settings.vue'), meta: { title: '设置' } },
  { path: '/about', name: 'about', component: () => import('@/views/About.vue'), meta: { title: '关于' } },
  // hover 悬浮卡片：bare 布局（无 Sidebar/TopBar）
  { path: '/hover', name: 'hover', component: () => import('@/views/HoverPopup.vue'), meta: { title: '悬浮卡片', bare: true } },
];

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
});
