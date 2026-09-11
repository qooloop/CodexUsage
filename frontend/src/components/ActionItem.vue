<script setup lang="ts">
/**
 * ActionItem - 快捷操作行（设计图：本会话/快捷操作 卡里的"立即刷新 Ctrl+R"）
 * icon | label（占主） | shortcut hint | arrow
 */
defineProps<{
  icon: 'refresh' | 'panel' | 'settings';
  label: string;
  shortcut?: string;
  arrow?: string; // 默认 "›"
  to?: string; // 点击跳转（router.push）
}>();
</script>

<template>
  <button type="button" class="action-item">
    <span class="ico">
      <svg v-if="icon === 'refresh'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12a9 9 0 1 1-3.5-7.1" />
        <path d="M21 3v6h-6" />
      </svg>
      <svg v-else-if="icon === 'panel'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="3" width="7" height="7" />
        <rect x="14" y="3" width="7" height="7" />
        <rect x="3" y="14" width="7" height="7" />
        <rect x="14" y="14" width="7" height="7" />
      </svg>
      <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />
      </svg>
    </span>
    <span class="label">{{ label }}</span>
    <span v-if="shortcut" class="shortcut tabular">{{ shortcut }}</span>
    <span v-if="!shortcut" class="arrow">{{ arrow ?? '›' }}</span>
  </button>
</template>

<style scoped>
.action-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  width: 100%;
  text-align: left;
  font-family: inherit;
  margin: 4px 0;
  font-size: var(--fs-13);
  border: 1px solid var(--border);
  background: rgba(139, 165, 186, 0.10);
  border-radius: 10px;
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
}
.action-item:disabled { opacity: .65; cursor: wait; }
.action-item:hover {
  background: var(--bg-hover);
  border-color: var(--border-strong);
}
.ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  color: var(--text-primary);
  flex-shrink: 0;
}
.label {
  flex: 1;
  color: var(--text-primary);
}
.shortcut {
  color: var(--text-tertiary);
  font-size: 10px;
  font-family: -apple-system, "SF Mono", Menlo, Consolas, monospace;
}
.arrow {
  color: var(--text-tertiary);
  font-size: 16px;
  line-height: 1;
  opacity: 0.6;
}
.action-item:hover .arrow {
  opacity: 1;
  color: var(--text-secondary);
}
</style>
