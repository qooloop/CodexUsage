<script setup lang="ts">
/**
 * Settings - 设置（需求文档 §51）
 * 持久化：后端 %APPDATA%\CodexUsageMonitor\config.json（Go 端自动应用开机启动/刷新间隔）
 * 浏览器调试环境回退 localStorage
 */
import { ref, watch, onMounted } from 'vue';
import {
  wailsGetConfig,
  wailsSaveConfig,
  isWails,
  type AppConfig,
} from '@/lib/wails';

const KEY = 'codex-monitor:settings:v1';
const defaults: AppConfig = {
  refreshInterval: 60,
  startWithWindows: false,
  hideOnStartup: true,
  showHoverPopup: true,
  trayMetric: 'shortQuota',
  showLongBar: true,
  notify5hBelow: 20,
  notify7dBelow: 20,
};

const s = ref<AppConfig>({ ...defaults });
let dirty = false;

onMounted(async () => {
  if (isWails()) {
    const c = await wailsGetConfig();
    if (c) s.value = { ...defaults, ...c };
  } else {
    try {
      const raw = localStorage.getItem(KEY);
      if (raw) s.value = { ...defaults, ...JSON.parse(raw) };
    } catch (e) {
      // ignore
    }
  }
  // 首次加载完成后再监听变更
  setTimeout(() => (dirty = true), 0);
});

watch(
  s,
  async (v) => {
    if (!dirty) return;
    if (isWails()) {
      const saved = await wailsSaveConfig({ ...v });
      if (saved && JSON.stringify(saved) !== JSON.stringify(v)) s.value = { ...v, ...saved };
    } else {
      try {
        localStorage.setItem(KEY, JSON.stringify(v));
      } catch (e) {
        // ignore
      }
    }
  },
  { deep: true },
);

const intervals: { v: number; label: string }[] = [
  { v: 5, label: '5 秒' },
  { v: 10, label: '10 秒' },
  { v: 30, label: '30 秒' },
  { v: 60, label: '1 分钟' },
  { v: 300, label: '5 分钟' },
];
const notifyOpts: { v: number; label: string }[] = [
  { v: 10, label: '10%' },
  { v: 20, label: '20%' },
  { v: 30, label: '30%' },
  { v: 0, label: '关闭' },
];
</script>

<template>
  <div class="settings">
    <h2 class="title">设置</h2>

    <section class="group">
      <div class="group-title">常规</div>
      <div class="card">
        <label class="row">
          <span>Windows 启动时自动运行</span>
          <input v-model="s.startWithWindows" type="checkbox" />
        </label>
        <label class="row">
          <span>启动后隐藏主窗口（仅托盘）</span>
          <input v-model="s.hideOnStartup" type="checkbox" />
        </label>
        <label class="row">
          <span>悬停托盘显示详情</span>
          <input v-model="s.showHoverPopup" type="checkbox" />
        </label>
      </div>
    </section>

    <section class="group">
      <div class="group-title">数据</div>
      <div class="card">
        <div class="row col">
          <div class="row-label">刷新间隔</div>
          <div class="seg">
            <button
              v-for="opt in intervals"
              :key="opt.v"
              :class="{ active: s.refreshInterval === opt.v }"
              @click="s.refreshInterval = opt.v"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <section class="group">
      <div class="group-title">托盘</div>
      <div class="card">
        <div class="row col">
          <div class="row-label">托盘主数字</div>
          <div class="seg">
            <button
              :class="{ active: s.trayMetric === 'shortQuota' }"
              @click="s.trayMetric = 'shortQuota'"
            >
              5 小时额度
            </button>
            <button
              :class="{ active: s.trayMetric === 'longQuota' }"
              @click="s.trayMetric = 'longQuota'"
            >
              7 天额度
            </button>
          </div>
        </div>
        <label class="row">
          <span>显示底部长期额度条</span>
          <input v-model="s.showLongBar" type="checkbox" />
        </label>
      </div>
    </section>

    <section class="group">
      <div class="group-title">提醒</div>
      <div class="card">
        <div class="row col">
          <div class="row-label">5 小时低于阈值提醒</div>
          <div class="seg">
            <button
              v-for="opt in notifyOpts"
              :key="'a' + opt.v"
              :class="{ active: s.notify5hBelow === opt.v }"
              @click="s.notify5hBelow = opt.v"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
        <div class="row col">
          <div class="row-label">7 天低于阈值提醒</div>
          <div class="seg">
            <button
              v-for="opt in notifyOpts"
              :key="'b' + opt.v"
              :class="{ active: s.notify7dBelow === opt.v }"
              @click="s.notify7dBelow = opt.v"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.settings {
  padding: 20px 24px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 720px;
  height: 100%;
  overflow: auto;
}
.title {
  font-size: var(--fs-18);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 4px;
  letter-spacing: 0.3px;
}
.group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.group-title {
  font-size: var(--fs-13);
  color: var(--text-tertiary);
  font-weight: 500;
  letter-spacing: 0.5px;
  padding: 0 4px;
}
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 4px 4px;
  box-shadow: var(--shadow-inner);
}
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  font-size: var(--fs-14);
  color: var(--text-primary);
}
.row.col {
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}
.row-label {
  color: var(--text-secondary);
  font-size: var(--fs-13);
}
.row input[type='checkbox'] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--accent);
}
.seg {
  display: flex;
  gap: 6px;
  background: rgba(255, 255, 255, 0.04);
  padding: 4px;
  border-radius: 10px;
  border: 1px solid var(--border);
}
.seg button {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: var(--fs-13);
  padding: 6px 12px;
  border-radius: 7px;
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
}
.seg button:hover {
  color: var(--text-primary);
}
.seg button.active {
  background: rgba(53, 230, 165, 0.15);
  color: var(--accent);
  font-weight: 500;
}
</style>
