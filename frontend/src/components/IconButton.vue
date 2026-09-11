<script setup lang="ts">
/**
 * IconButton - 窗口右上角控制按钮（— □ ×）
 * 复用，hover 时背景变色
 */
defineProps<{
  variant?: 'min' | 'max' | 'close';
}>();
</script>

<template>
  <button class="win-btn" :class="variant ?? 'min'">
    <svg v-if="variant === 'min'" width="12" height="12" viewBox="0 0 12 12">
      <rect x="2" y="5.5" width="8" height="1.5" rx="0.75" fill="currentColor" />
    </svg>
    <svg v-else-if="variant === 'max'" width="12" height="12" viewBox="0 0 12 12">
      <rect x="2.5" y="2.5" width="7" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.2" />
    </svg>
    <svg v-else width="12" height="12" viewBox="0 0 12 12">
      <path d="M3 3 L9 9 M9 3 L3 9" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
    </svg>
  </button>
</template>

<style scoped>
.win-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 6px;
  color: var(--text-secondary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
}
.win-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}
.win-btn.close:hover {
  background: var(--danger);
  color: #fff;
}
</style>
