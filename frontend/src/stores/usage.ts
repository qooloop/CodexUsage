// src/stores/usage.ts
// 订阅 Wails 'snapshots:update' 事件；也兼容纯浏览器调试（wails runtime 不存在时走 mock）

import { quotaPercent } from '@/lib/format';
import { defineStore } from 'pinia';
import type { Snapshot } from '@/types/usage';

interface UsageState {
  refreshing: boolean;
  snapshots: Snapshot[];
  lastEventAt: number;
  isMock: boolean; // 非 Wails 环境下用 mock 数据
}

export const useUsageStore = defineStore('usage', {
  state: (): UsageState => ({
    refreshing: false,
    snapshots: [],
    lastEventAt: 0,
    isMock: false,
  }),

  getters: {
    /** 取第一个（也是当前唯一）provider 快照 */
    primary(state): Snapshot | null {
      return state.snapshots[0] ?? null;
    },
    syncLabel(): string {
      const s = this.primary;
      if (s?.quotaSource === 'local') return s.source === 'fresh' ? '本地已同步' : '本地记录 · 待更新';
      return s?.error ? '刷新失败 · 缓存' : s?.source === 'fresh' ? '同步正常' : '等待更新';
    },
    syncColor(): 'good' | 'warn' | 'idle' {
      const s = this.primary;
      if (s?.source === 'fresh') return 'good';
      return s?.error ? 'warn' : 'idle';
    },
    /** 给 UI 用的格式化字段 */
    quota5hPct(): number {
      const p = this.primary?.primary;
      return p ? quotaPercent(p.remainingPct) : 0;
    },
    quota7dPct(): number {
      const p = this.primary?.secondary;
      return p ? quotaPercent(p.remainingPct) : 0;
    },
    todayTokens(): number {
      return this.primary?.token.today ?? 0;
    },
    sessionTokens(): number {
      return this.primary?.token.currentTask ?? 0;
    },
    cacheHitRate(): number {
      return this.primary?.token.cacheHitRate ?? 0;
    },
  },

  actions: {
    setSnapshots(list: Snapshot[]) {
      this.snapshots = list ?? [];
      this.lastEventAt = Date.now();
    },
  },
});
