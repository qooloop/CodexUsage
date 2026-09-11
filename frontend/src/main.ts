// src/main.ts - Vue 3 入口
// 1. 挂载 Pinia（状态：snapshots / theme / settings）
// 2. 挂载 Vue Router（4 个视图：总览 / 统计 / 设置 / 关于）
// 3. 引入设计 token + 全局样式
// 4. Wails 环境检测：仅在生产环境（wails runtime 存在）下才订阅事件

import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { router } from './router';

import './styles/tokens.css';
import './styles/global.css';

const app = createApp(App);
app.use(createPinia());
app.use(router);
app.mount('#app');
