<script setup lang="ts">
/**
 * App.vue - 根组件
 * - /hover 路由 → bare 布局（仅悬浮卡片，设计图右侧卡片）
 * - 其余路由 → Sidebar + TopBar + router-view（详细面板，设计图左侧面板）
 * - 监听后端 ui:mode 事件切换视图；订阅 snapshots:update 回灌 store
 */
import { onMounted, onBeforeUnmount, computed, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Sidebar from './components/Sidebar.vue';
import TopBar from './components/TopBar.vue';
import { useUsageStore } from './stores/usage';
import { wailsGetSnapshots, wailsOnSnapshotsUpdate, wailsOnUIMode, wailsRefresh, wailsHidePopup, wailsOpenOfficialPanel, wailsGetUIMode, wailsConfirmUIMode } from './lib/wails';

const store = useUsageStore();
const route = useRoute();
const router = useRouter();

let unsubRefresh: (() => void) | undefined;
let unsubscribe: (() => void) | null = null;
let unsubMode: (() => void) | null = null;

const isBare = computed(() => Boolean(route.meta.bare));
const viewName = computed(() => String(route.name ?? 'overview'));

onMounted(async () => {
  unsubRefresh = window.runtime?.EventsOn('refresh:state', (value) => { store.refreshing = Boolean(value); });
  // 拉一次 + 订阅事件
  try {
    const initial = await wailsGetSnapshots();
    store.setSnapshots(initial);
  } catch (e) {
    // ignore
  }
  unsubscribe = wailsOnSnapshotsUpdate((snaps) => {
    store.setSnapshots(snaps);
  });

  // Recover a mode requested before the WebView event subscription was ready.
  unsubMode = wailsOnUIMode(applyMode);
  const initialMode = await wailsGetUIMode();
  if (initialMode) await applyMode(initialMode);

  // 键盘快捷键
  window.addEventListener('keydown', onKeydown);
});

async function applyMode(mode: string) {
 if (!['hover', 'overview', 'settings', 'stats', 'about'].includes(mode)) return;
 await router.replace({ name: mode });
 await nextTick();
 await wailsConfirmUIMode(mode);
}

onBeforeUnmount(() => {
  unsubRefresh?.();
  unsubscribe?.();
  unsubMode?.();
  window.removeEventListener('keydown', onKeydown);
});

async function onKeydown(e: KeyboardEvent) {
  if (e.key === 'F5' || (e.ctrlKey && e.key.toLowerCase() === 'r')) {
    e.preventDefault();
    const snaps = await wailsRefresh();
    store.setSnapshots(snaps);
  } else if (e.ctrlKey && e.key.toLowerCase() === 'd') {
 e.preventDefault(); wailsOpenOfficialPanel();
 } else if (e.key === 'Escape') {
    wailsHidePopup();
  }
}
</script>

<template>
  <!-- hover 悬浮卡片：裸布局 -->
  <div v-if="isBare" class="app-shell app-shell--hover">
    <router-view />
  </div>

  <!-- 详细面板：完整布局（设计图：顶栏全宽贯穿，侧栏在顶栏下方） -->
  <div v-else class="app-shell">
    <TopBar />
    <div class="layout">
      <Sidebar />
      <main class="main">
        <div class="content">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" :key="viewName" />
            </transition>
          </router-view>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
}
.layout {
  flex: 1;
  min-height: 0;
  display: flex;
  position: relative;
  z-index: 1;
}
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100%;
}
.content {
  flex: 1;
  overflow: auto;
  padding: 0;
}
.app-shell--hover {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
