// src/lib/format.ts
// 统一格式化函数 —— 数字、时间、百分比

/** 1234 -> "1,234"；1234567 -> "1,234,567" */
export function fmtInt(n: number | null | undefined): string {
  if (n == null || isNaN(n)) return '—';
  return Math.round(n).toLocaleString('en-US');
}

/**
 * 中文大数：
 *   >= 1e8 -> "x.xx亿"
 *   >= 1e4 -> "x万"（保留 0~1 位小数）
 *   < 1e4  -> 整数（千分位）
 */
export function fmtBig(n: number | null | undefined): string {
  if (n == null || isNaN(n)) return '—';
  const v = Math.round(n);
  if (v >= 1_0000_0000) return (v / 1_0000_0000).toFixed(2) + '亿';
  if (v >= 1_0000) {
    // >=100万 取整；<100万 保留 1 位但去掉尾零（210040 -> 21万，235000 -> 23.5万）
    if (v >= 100_0000) return Math.round(v / 1_0000) + '万';
    const s = (v / 1_0000).toFixed(1);
    return (s.endsWith('.0') ? s.slice(0, -2) : s) + '万';
  }
  return v.toLocaleString('en-US');
}

/** 缓存命中率 0~100 -> "97.2%" */
export function fmtPct(p: number | null | undefined, frac = 1): string {
  if (p == null || isNaN(p)) return '—';
  return p.toFixed(frac) + '%';
}

/** unix 秒 -> "MM-DD HH:mm"（本地时区） */
export function fmtResetAbs(epochSec: number | null | undefined): string {
  if (!epochSec) return '—';
  const d = new Date(epochSec * 1000);
  const pad = (x: number) => String(x).padStart(2, '0');
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** unix 秒 -> 当天 "13:46"，跨天 "09-15 12:21"（设计图口径：不加"今天"前缀） */
export function fmtResetSmart(epochSec: number | null | undefined): string {
  if (!epochSec) return '—';
  const d = new Date(epochSec * 1000);
  const now = new Date();
  const pad = (x: number) => String(x).padStart(2, '0');
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate();
  if (sameDay) {
    return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** unix 秒 距今 -> "2h13m" / "13m" / "已重置" */
export function fmtResetCountdown(epochSec: number | null | undefined): string {
  if (!epochSec) return '—';
  const ms = epochSec * 1000 - Date.now();
  if (ms <= 0) return '已重置';
  const m = Math.floor(ms / 60000);
  if (m < 1) return '<1m';
  if (m < 60) return `${m}m`;
  if (m < 1440) return `${Math.floor(m / 60)}h${m % 60}m`;
  return `${Math.floor(m / 1440)}d${Math.floor((m % 1440) / 60)}h`;
}

/** ms -> "HH:MM:SS" */
export function fmtClock(ms: number | null | undefined): string {
  if (!ms) return '—';
  const d = new Date(ms);
  const pad = (x: number) => String(x).padStart(2, '0');
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

/** ms -> "HH:MM" */
export function fmtClockShort(ms: number | null | undefined): string {
  if (!ms) return '—';
  const d = new Date(ms);
  const pad = (x: number) => String(x).padStart(2, '0');
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** 0~100 -> 健康色 class（绿/黄/红） */
export function quotaColor(p: number | null | undefined): 'good' | 'warn' | 'bad' {
  if (p == null || isNaN(p)) return 'good';
  if (p >= 50) return 'good';
  if (p >= 20) return 'warn';
  return 'bad';
}

/** Remaining quota uses truncation, never rounding up (57.9% -> 57%). */
export function quotaPercent(value: number): number {
  return Number.isFinite(value) ? Math.floor(Math.max(0, Math.min(100, value))) : 0;
}
export function fmtQuota(value: number | null | undefined): string {
  return value == null || !Number.isFinite(value) ? '—' : `${quotaPercent(value)}%`;
}
