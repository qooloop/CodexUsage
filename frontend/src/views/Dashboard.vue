<script setup lang="ts">
/**
 * Dashboard - 总览页（设计图 1:1 复刻）
 *
 * 区域：
 *   1. 配额使用（3 张卡：5小时 / 7天 / 今日 Token）
 *   2. 本会话 + 快捷操作（两栏）
 *   3. 底部 footer
 */
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import QuotaCard from '@/components/QuotaCard.vue';
import StatItem from '@/components/StatItem.vue';
import ActionItem from '@/components/ActionItem.vue';
import { useUsageStore } from '@/stores/usage';
import { fmtBig, fmtPct, fmtResetSmart, fmtClockShort, quotaColor } from '@/lib/format';
import { wailsRefresh, wailsOpenCodexHome, wailsOpenOfficialPanel } from '@/lib/wails';

const store = useUsageStore();
const router = useRouter();

const s = computed(() => store.primary);
const quota5h = computed(() => s.value?.primary ?? null);
const quota7d = computed(() => s.value?.secondary ?? null);

const fill5h = computed(() => Math.max(0, Math.min(100, quota5h.value?.remainingPct ?? 0)));
const fill7d = computed(() => Math.max(0, Math.min(100, quota7d.value?.remainingPct ?? 0)));
// Token has no subscription quota: show its share of the last seven days.
const todayFill = computed(() => {
 const t = s.value?.token;
 return t?.near7d ? Math.max(0, Math.min(100, t.today / t.near7d * 100)) : 0;
});

async function onRefresh() {
  const snaps = await wailsRefresh();
  store.setSnapshots(snaps);
}

function onOpenHome() {
  wailsOpenCodexHome();
}

function onOpenPanel() {
  wailsOpenOfficialPanel();
}

function onGoSettings() {
  router.push({ name: 'settings' });
}
</script>

<template>
  <div class="dashboard">
    <!-- 配额使用 3 卡片 -->
    <section class="block">
      <div class="block-title">配额使用</div>
      <div class="cards-3">
        <QuotaCard
          icon="clock"
          label="5小时额度"
          :value="quota5h?.remainingPct ?? null"
          :fill-percent="fill5h"
          :fill-color="quotaColor(quota5h?.remainingPct)"
          :foot-left="'重置 ' + fmtResetSmart(quota5h?.resetsAt ?? null)"
        />
        <QuotaCard
          icon="calendar"
          label="7天额度"
          :value="quota7d?.remainingPct ?? null"
          :fill-percent="fill7d"
          :fill-color="quotaColor(quota7d?.remainingPct)"
          :foot-left="'重置 ' + fmtResetSmart(quota7d?.resetsAt ?? null)"
        />
        <QuotaCard
          icon="token"
          title="进度条表示今日 Token 占近 7 天 Token 的比例"
          label="今日 Token"
          :value="s?.token.today ?? null"
          :fill-percent="todayFill"
          fill-color="blue"
          :foot-left="'上次更新 ' + fmtClockShort(s?.ts)"
        />
      </div>
    </section>

    <!-- 本会话 + 快捷操作 -->
    <section class="block two-col">
      <div class="card panel">
        <div class="panel-title">本会话</div>
        <div class="panel-list">
          <StatItem icon="token" label="会话累计 Token" :value="fmtBig(s?.token.currentTask ?? 0)" />
          <StatItem icon="cache" label="缓存命中率" :value="fmtPct(s?.token.cacheHitRate, 1)" />
          <StatItem icon="request" label="最近一次请求" :value="fmtClockShort(s?.lastRequestAt)" />
          <StatItem
            icon="status"
            label="当前状态"
            :title="s?.error || ''"
            :value="store.syncLabel"
            :status-color="store.syncColor"
            hint="?"
          />
        </div>
      </div>

      <div class="card panel">
        <div class="panel-title">快捷操作</div>
        <div class="panel-list">
          <ActionItem icon="refresh" :label="store.refreshing ? '正在刷新…' : '立即刷新'" :disabled="store.refreshing" shortcut="Ctrl + R" @click="onRefresh" />
          <ActionItem icon="panel" label="打开官方面板" shortcut="Ctrl + D" @click="onOpenPanel" />
          <ActionItem icon="settings" label="监控设置" @click="onGoSettings" />
        </div>
      </div>
    </section>

    <!-- 底部 -->
    <footer class="foot">
      <span class="foot-quote">" 更专注的 AI 工作体验 "</span>
      <span class="foot-meta">Codex 用量监控 · 让配额尽在掌握 <span class="heart">❤</span></span>
    </footer>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 14px 20px 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}
.block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.block-title {
  font-size: var(--fs-16);
  color: var(--text-primary);
  font-weight: 600;
  letter-spacing: 0.5px;
}
.cards-3 {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 12px;
}
.two-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  flex: 1;
  min-height: 0;
}
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 10px 14px 8px;
  box-shadow: var(--shadow-inner);
}
.panel {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.panel-title {
  font-size: var(--fs-16);
  color: var(--text-primary);
  font-weight: 600;
  padding: 4px 4px 2px;
}
.panel-list {
  display: flex;
  flex-direction: column;
}
.foot {
  margin-top: auto;
  padding: 8px 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 10px;
  color: var(--text-tertiary);
}
.foot-quote {
  font-style: italic;
}
.foot-meta {
  color: var(--text-tertiary);
}
.heart {
  color: #60758b;
  margin-left: 2px;
}
</style>
