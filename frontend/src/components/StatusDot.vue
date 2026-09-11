<script setup lang="ts">
/**
 * StatusDot - 状态圆点（绿/红/灰）
 * 设计图里 sidebar 底部 ● 运行中、卡片底 ● 同步正常 都用这个
 */
defineProps<{
  color?: 'good' | 'warn' | 'bad' | 'idle';
  pulse?: boolean;
}>();
</script>

<template>
  <span class="dot" :class="[color ?? 'good', { pulse: pulse ?? false }]"></span>
</template>

<style scoped>
.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.dot.good {
  background: var(--accent);
  box-shadow: 0 0 8px rgba(53, 230, 165, 0.55);
}
.dot.warn {
  background: var(--warning);
  box-shadow: 0 0 8px rgba(246, 196, 83, 0.5);
}
.dot.bad {
  background: var(--danger);
  box-shadow: 0 0 8px rgba(255, 97, 116, 0.5);
}
.dot.idle {
  background: var(--text-tertiary);
}
.pulse {
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
