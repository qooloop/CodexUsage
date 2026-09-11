// src/types/usage.ts
// 与 Go 端 Snapshot 结构对齐，由 Wails 自动生成的 wailsjs/go/models.ts 也有同款
// 这里重写一份便于前端强类型 & 单测

export interface QuotaWindow {
  usedPct: number;
  remainingPct: number;
  windowMinutes: number;
  resetsAt: number | null; // unix 秒
}

export interface TokenStats {
  today: number;
  near7d: number;
  total: number;
  currentTask: number;
  lastTurn: number;
  cacheHitRate: number;
}

export interface CreditsInfo {
  hasCredits: boolean;
  unlimited: boolean;
  balance: string;
}

export interface Snapshot {
  id: string;
  name: string;
  shortName: string;
  color: string;
  emoji: string;
  primary: QuotaWindow | null;
  secondary: QuotaWindow | null;
  token: TokenStats;
  credits: CreditsInfo | null;
  status: string; // "执行中" | "空闲"
  source: string; // "fresh" | "stale"
  quotaSource?: 'local' | 'online' | 'none';
  lastRequestAt?: number; // latest token event, ms
  ts: number; // quota snapshot, ms
  error: string;
}

export type ViewName = 'overview' | 'stats' | 'settings' | 'about';
