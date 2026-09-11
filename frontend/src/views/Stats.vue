<script setup lang="ts">
/**
 * Stats - 统计详情
 * 把当前 Snapshots 数据做时间分布、Token 分类、本次任务详情
 */
import { computed } from 'vue';
import { useUsageStore } from '@/stores/usage';
import StatItem from '@/components/StatItem.vue';
import { fmtBig, fmtInt, fmtPct, fmtClockShort } from '@/lib/format';
import { fmtQuota } from '@/lib/format';

const store = useUsageStore();
const s = computed(() => store.primary);

// 本次任务统计
const threadStats = computed(() => {
  const t = s.value?.token;
  if (!t) return null;
  return {
    currentTask: t.currentTask,
    lastTurn: t.lastTurn,
    today: t.today,
    near7d: t.near7d,
    total: t.total,
  };
});
</script>

<template>
  <div class="stats">
    <h2 class="title">统计详情</h2>

    <div v-if="!s" class="empty">暂无数据 — 请确认 Codex Desktop 已运行</div>

    <template v-else>
      <div class="grid">
        <div class="card panel">
          <div class="panel-title">本次任务</div>
          <div class="panel-list">
            <StatItem icon="token" label="任务累计" :value="fmtBig(threadStats?.currentTask ?? 0)" />
            <StatItem icon="cache" label="上次回复消耗" :value="fmtInt(threadStats?.lastTurn ?? 0)" />
          </div>
        </div>

        <div class="card panel">
          <div class="panel-title">时间分布</div>
          <div class="panel-list">
            <StatItem icon="request" label="今日 Token" :value="fmtBig(threadStats?.today ?? 0)" />
            <StatItem icon="request" label="近 7 天" :value="fmtBig(threadStats?.near7d ?? 0)" />
            <StatItem icon="token" label="累计（终身）" :value="fmtBig(threadStats?.total ?? 0)" />
          </div>
        </div>

        <div class="card panel">
          <div class="panel-title">效率指标</div>
          <div class="panel-list">
            <StatItem icon="cache" label="缓存命中率" :value="fmtPct(s.token.cacheHitRate, 1)" />
            <StatItem icon="time" label="最近数据时间" :value="fmtClockShort(s.ts)" />
            <StatItem
              icon="status"
              label="数据状态"
              :title="s.error || ''"
              :value="store.syncLabel"
              :status-color="s.source === 'fresh' ? 'good' : 'warn'"
            />
          </div>
        </div>

        <div class="card panel">
          <div class="panel-title">额度窗口</div>
          <div class="panel-list">
            <StatItem icon="clock" label="5h 剩余" :value="s.primary ? fmtQuota(s.primary.remainingPct) : '—'" />
            <StatItem icon="calendar" label="7d 剩余" :value="s.secondary ? fmtQuota(s.secondary.remainingPct) : '—'" />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.stats {
  padding: 20px 24px 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: 100%;
}
.title {
  font-size: var(--fs-18);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 4px;
  letter-spacing: 0.3px;
}
.empty {
  color: var(--text-tertiary);
  text-align: center;
  padding: 60px 0;
  font-size: var(--fs-13);
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 12px 16px 4px;
  box-shadow: var(--shadow-inner);
}
.panel {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.panel-title {
  font-size: var(--fs-13);
  color: var(--text-secondary);
  font-weight: 500;
  padding: 4px 4px 2px;
}
.panel-list {
  display: flex;
  flex-direction: column;
}
</style>
