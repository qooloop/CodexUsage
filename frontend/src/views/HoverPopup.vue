<script setup lang="ts">
/**
 * HoverPopup.vue - 悬停卡片（设计图右侧卡片 1:1 复刻，需求文档 §11）
 *
 * 结构：
 *   C Codex            ⚙
 *   ◷ 5小时额度   73%
 *     ▬▬▬▬▬▬▬     重置 13:46
 *   ▣ 7天额度     60%
 *     ▬▬▬▬▬▬▬     重置 09-15 12:21
 *   ◎ 今日 Token      21万
 *   ▮ 会话累计        714万
 *   ─────────────────────
 *   ● 同步正常    11:51:23
 */
import { computed } from 'vue';
import { useUsageStore } from '@/stores/usage';
import StatusDot from '@/components/StatusDot.vue';
import CodexLogo from '@/components/CodexLogo.vue';
import { fmtBig, fmtClock, fmtResetSmart, quotaColor } from '@/lib/format';
import { fmtQuota } from '@/lib/format';
import { wailsOpenDashboard } from '@/lib/wails';

const store = useUsageStore();
const s = computed(() => store.primary);

const fill5h = computed(() => clamp(s.value?.primary?.remainingPct ?? 0));
const fill7d = computed(() => clamp(s.value?.secondary?.remainingPct ?? 0));

function clamp(v: number) {
  return Math.max(0, Math.min(100, v));
}
</script>

<template>
  <div class="hover-pop">
    <!-- 头部：C Codex ⚙ -->
    <header class="hp-head">
      <div class="hp-brand">
        <CodexLogo :size="22" />
        <span class="hp-name">Codex</span>
      </div>
      <button class="hp-gear" title="设置" @click="wailsOpenDashboard('settings')">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="3" />
          <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />
        </svg>
      </button>
    </header>

    <!-- 5小时额度 -->
    <section class="hp-quota">
      <div class="hp-row">
        <span class="hp-ico clock">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="9" />
            <polyline points="12 7 12 12 15 14" />
          </svg>
        </span>
        <span class="hp-label">5小时额度</span>
        <span class="hp-pct tabular">{{ s?.primary ? fmtQuota(s.primary.remainingPct) : '—' }}</span>
      </div>
      <div class="hp-bar-row">
        <div class="hp-bar">
          <div class="hp-bar-fill" :class="quotaColor(s?.primary?.remainingPct)" :style="{ width: fill5h + '%' }"></div>
        </div>
        <span class="hp-reset">重置 {{ fmtResetSmart(s?.primary?.resetsAt ?? null) }}</span>
      </div>
    </section>

    <!-- 7天额度 -->
    <section class="hp-quota">
      <div class="hp-row">
        <span class="hp-ico cal">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="4" width="18" height="18" rx="2" />
            <line x1="3" y1="10" x2="21" y2="10" />
            <line x1="8" y1="2" x2="8" y2="6" />
            <line x1="16" y1="2" x2="16" y2="6" />
          </svg>
        </span>
        <span class="hp-label">7天额度</span>
        <span class="hp-pct tabular">{{ s?.secondary ? fmtQuota(s.secondary.remainingPct) : '—' }}</span>
      </div>
      <div class="hp-bar-row">
        <div class="hp-bar">
          <div class="hp-bar-fill" :class="quotaColor(s?.secondary?.remainingPct)" :style="{ width: fill7d + '%' }"></div>
        </div>
        <span class="hp-reset">重置 {{ fmtResetSmart(s?.secondary?.resetsAt ?? null) }}</span>
      </div>
    </section>

    <!-- Token 简报 -->
    <section class="hp-stats">
      <div class="hp-row">
        <span class="hp-ico token">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <ellipse cx="12" cy="6" rx="8" ry="3" />
            <path d="M4 6v12c0 1.66 3.58 3 8 3s8-1.34 8-3V6" />
            <path d="M4 12c0 1.66 3.58 3 8 3s8-1.34 8-3" />
          </svg>
        </span>
        <span class="hp-label">今日 Token</span>
        <span class="hp-pct tabular">{{ fmtBig(s?.token.today ?? 0) }}</span>
      </div>
      <div class="hp-row">
        <span class="hp-ico bars">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 20h18M5 20V10M10 20V4M15 20v-8M20 20V14" />
          </svg>
        </span>
        <span class="hp-label">会话累计</span>
        <span class="hp-pct tabular">{{ fmtBig(s?.token.currentTask ?? 0) }}</span>
      </div>
    </section>

    <!-- 底部状态 -->
    <footer class="hp-foot">
      <span class="hp-status" :class="{ 'is-stale': s?.source !== 'fresh' }" :title="s?.error || (s?.source !== 'fresh' ? '当前显示最近记录，等待刷新确认' : '')">
        <StatusDot :color="store.syncColor" :pulse="s?.source === 'fresh'" />
        <span>{{ store.refreshing ? '正在刷新…' : store.syncLabel }}</span>
      </span>
      <span class="hp-time tabular" title="当前额度记录时间">{{ fmtClock(s?.ts) }}</span>
    </footer>
  </div>
</template>

<style scoped>
.hover-pop { width:100%; height:100%; padding:14px 16px 12px; display:flex; flex-direction:column; background:linear-gradient(145deg,#14253b,#0b1521 75%,#13242c); }
.hp-head,.hp-brand,.hp-foot,.hp-status { display:flex; align-items:center; }
.hp-head { justify-content:space-between; padding-bottom:12px; border-bottom:1px solid var(--border); }
.hp-brand { gap:10px; } .hp-name { font-size:16px; font-weight:600; }
.hp-gear { display:grid; place-items:center; width:26px; height:26px; border:0; border-radius:6px; background:transparent; color:var(--text-secondary); cursor:pointer; }
.hp-gear:hover { color:white; background:var(--bg-hover); }
.hp-quota { position:relative; height:49px; padding-top:10px; border-bottom:1px solid var(--border); }
.hp-row { display:grid; grid-template-columns:22px 78px 1fr; align-items:center; gap:10px; }
.hp-ico { display:flex; color:#51e4cf; } .hp-ico svg { width:19px; height:19px; }
.hp-ico.token { color:#a8cce9; } .hp-ico.bars { color:var(--text-secondary); }
.hp-label { font-size:13px; color:var(--text-secondary); white-space:nowrap; }
.hp-pct { font-size:14px; font-weight:600; }
.hp-bar-row { margin:4px 0 0 120px; }
.hp-bar { height:7px; border-radius:99px; background:#223448; overflow:hidden; }
.hp-bar-fill { height:100%; border-radius:99px; transition:width 300ms ease-out; }
.hp-bar-fill.good { background:linear-gradient(90deg,#31ef78,#12e8bb); }
.hp-bar-fill.warn { background:var(--warning); } .hp-bar-fill.bad { background:var(--danger); }
.hp-reset { position:absolute; right:0; top:11px; font-size:11px; color:var(--text-secondary); }
.hp-stats .hp-row { height:40px; border-bottom:1px solid var(--border); }
.hp-stats .hp-row:last-child { border-bottom:0; }
.hp-foot { margin-top:auto; padding-top:12px; border-top:1px solid var(--border); justify-content:space-between; font-size:13px; }
.hp-status { gap:10px; color:var(--accent); } .hp-time { color:var(--text-secondary); }
.hp-status.is-stale { color:var(--warning); }
</style>
