<script setup lang="ts">
/**
 * TopBar - 顶部 brand + 窗口控制
 * 设计图：左上 C 蓝色单环 + "Codex 用量监控" / "实时监控 · 配额可视化 · 常驻托盘"
 * 右上：— □ ×
 */
import CodexLogo from './CodexLogo.vue';
import IconButton from './IconButton.vue';
import { wailsHidePopup, wailsToggleMaximize } from '@/lib/wails';
</script>

<template>
  <header class="topbar">
    <div class="brand">
      <div class="logo">
        <CodexLogo :size="34" />
      </div>
      <div class="text">
        <div class="title">Codex 用量监控</div>
        <div class="subtitle">实时监控 · 配额可视化 · 常驻托盘</div>
      </div>
    </div>

    <div class="win-controls">
      <IconButton variant="min" @click="wailsHidePopup" title="最小化到托盘" />
      <IconButton variant="max" @click="wailsToggleMaximize" title="最大化 / 还原" />
      <IconButton variant="close" @click="wailsHidePopup" title="关闭（隐藏到托盘）" />
    </div>
  </header>
</template>

<style scoped>
.topbar {
  height: 73px;
  padding: 0 16px 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  background: var(--bg-sidebar);
  flex-shrink: 0;
  /* 无边框窗口：顶栏即拖动把手 */
  --wails-draggable: drag;
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}
.logo {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: transparent;
}
.text {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.title {
  font-size: var(--fs-20);
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.1;
  letter-spacing: 0.3px;
}
.subtitle {
  font-size: var(--fs-12);
  color: var(--text-secondary);
  letter-spacing: 0.2px;
}
.win-controls {
  display: flex;
  gap: 4px;
  /* 按钮不可拖拽，否则点不到 */
  --wails-draggable: no-drag;
}
</style>