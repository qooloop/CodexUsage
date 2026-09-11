<script setup lang="ts">
/**
 * Sidebar - 左侧导航
 * 设计图：总览（激活）/ 统计详情 / 设置 / 关于 + 底部 v1.0.0 + 状态
 */
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import StatusDot from './StatusDot.vue';
import { useUsageStore } from '@/stores/usage';

import { wailsGetConfig } from '@/lib/wails';
const autostart = ref(false);
let offConfig: (() => void) | undefined;
onMounted(async () => { autostart.value = (await wailsGetConfig())?.startWithWindows ?? false; offConfig = window.runtime?.EventsOn('config:update', c => { autostart.value = c.startWithWindows; }); });
onBeforeUnmount(() => offConfig?.());
const route = useRoute();
const router = useRouter();
const store = useUsageStore();

const items = [
  { name: 'overview', label: '总览', icon: 'home' as const },
  { name: 'stats', label: '统计详情', icon: 'chart' as const },
  { name: 'settings', label: '设置', icon: 'gear' as const },
  { name: 'about', label: '关于', icon: 'info' as const },
];

const activeName = computed(() => String(route.name ?? 'overview'));
const isOnline = computed(() => store.primary?.source === 'fresh');
const statusText = computed(() => (isOnline.value ? '运行中' : '已离线'));
</script>

<template>
  <aside class="sidebar">
    <nav class="nav">
      <button
        v-for="item in items"
        :key="item.name"
        class="nav-item"
        :class="{ active: activeName === item.name }"
        @click="router.push({ name: item.name })"
      >
        <span class="ico">
          <svg v-if="item.icon === 'home'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 10.5 12 3l9 7.5" />
            <path d="M5 9.5V21h14V9.5" />
            <path d="M10 21v-6h4v6" />
          </svg>
          <svg v-else-if="item.icon === 'chart'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 3v18h18" />
            <rect x="7" y="13" width="3" height="5" rx="0.5" />
            <rect x="12" y="9" width="3" height="9" rx="0.5" />
            <rect x="17" y="5" width="3" height="13" rx="0.5" />
          </svg>
          <svg v-else-if="item.icon === 'gear'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3" />
            <path d="M12 2v3M12 19v3M4.2 4.2l2.2 2.2M17.6 17.6l2.2 2.2M2 12h3M19 12h3M4.2 19.8l2.2-2.2M17.6 6.4l2.2-2.2" />
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="9" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <circle cx="12" cy="16" r="0.8" fill="currentColor" />
          </svg>
        </span>
        <span class="label">{{ item.label }}</span>
      </button>
    </nav>

    <div class="meta">
      <div class="version">v1.0.0</div>
      <div class="status">
        <StatusDot :color="isOnline ? 'good' : 'idle'" :pulse="isOnline" />
        <span>{{ statusText }}</span>
      </div>
      <div class="boot">{{ autostart ? '开机自动启动' : '开机启动已关闭' }}</div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 146px;
  height: 100%;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 18px 6px 16px;
  flex-shrink: 0;
}
.nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border: none;
  background: transparent;
  border-radius: var(--radius-btn);
  color: var(--text-secondary);
  font-size: var(--fs-14);
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
  text-align: left;
}
.nav-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-primary);
}
.nav-item.active {
  background: rgba(53, 206, 210, 0.13);
  box-shadow: inset 3px 0 #35e6a5;
  color: #64e2ee;
}
.nav-item.active .ico {
  color: #64e2ee;
}
.ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  color: var(--text-tertiary);
}
.nav-item:hover .ico {
  color: var(--text-secondary);
}
.label {
  font-size: var(--fs-14);
}

.meta {
  margin-top: auto;
  padding: 12px 12px 4px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: var(--fs-12);
  color: var(--text-tertiary);
}
.version {
  font-size: var(--fs-12);
  color: var(--text-tertiary);
}
.status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary);
}
.boot {
  color: var(--text-tertiary);
  font-size: var(--fs-12);
}
</style>
