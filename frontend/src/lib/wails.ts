// src/lib/wails.ts
// 前端与 Wails 后端的桥接
// - 生产环境（wails runtime 存在）：调真实后端
// - 浏览器开发环境（无 wails runtime）：返回 mock 数据，便于纯前端调试

import type { Snapshot } from '@/types/usage';

export interface AppConfig {
  refreshInterval: number;
  startWithWindows: boolean;
  hideOnStartup: boolean;
  showHoverPopup: boolean;
  trayMetric: 'shortQuota' | 'longQuota';
  showLongBar: boolean;
  notify5hBelow: number;
  notify7dBelow: number;
}

type WailsRuntime = {
  EventsOn: (event: string, cb: (data: any) => void) => () => void;
};

// 与 Go 端 Config json tag 对齐
interface GoConfig {
  refreshInterval: number;
  startWithWindows: boolean;
  hideOnStartup: boolean;
  showHoverPopup: boolean;
  trayMetric: 'shortQuota' | 'longQuota';
  showLongBar: boolean;
  notify5hBelow: number;
  notify7dBelow: number;
}

declare global {
  interface Window {
    runtime?: WailsRuntime;
    go?: {
      desktop?: {
        App?: {
          GetUIMode?: () => Promise<string>;
          ConfirmUIMode?: (route: string) => Promise<void>;
          GetSnapshots?: () => Promise<Snapshot[]>;
          Refresh?: () => Promise<Snapshot[]>;
          OpenOfficialPanel?: () => Promise<void>;
          OpenCodexHome?: () => Promise<string>;
          HidePopup?: () => Promise<void>;
          Quit?: () => Promise<void>;
          ToggleMaximize?: () => Promise<void>;
          GetConfig?: () => Promise<GoConfig>;
          SaveConfig?: (c: GoConfig) => Promise<GoConfig>;
          OpenDashboard?: (route: string) => Promise<void>;
        };
      };
    };
  }
}

export const isWails = (): boolean => {
  return typeof window !== 'undefined' && !!window.runtime && !!window.go?.desktop?.App?.GetSnapshots;
};

const getApp = () => window.go!.desktop!.App!;

export const wailsGetSnapshots = async (): Promise<Snapshot[]> => {
  if (!isWails()) return mockSnapshots();
  return (await getApp().GetSnapshots!()) ?? [];
};

export const wailsRefresh = async (): Promise<Snapshot[]> => {
  if (!isWails()) return mockSnapshots();
  return (await getApp().Refresh!()) ?? [];
};

export const wailsOpenCodexHome = async (): Promise<string> => {
  if (!isWails()) return '';
  return (await getApp().OpenCodexHome!()) ?? '';
};

export const wailsHidePopup = async (): Promise<void> => {
  if (!isWails()) return;
  await getApp().HidePopup?.();
};

export const wailsQuit = async (): Promise<void> => {
  if (!isWails()) return;
  await getApp().Quit?.();
};

export const wailsToggleMaximize = async (): Promise<void> => {
  if (!isWails()) return;
  await getApp().ToggleMaximize?.();
};

// ---- 配置 ----

function goToAppConfig(c: GoConfig): AppConfig {
  return { ...c };
}

export const wailsGetConfig = async (): Promise<AppConfig | null> => {
  if (!isWails()) return null;
  const c = await getApp().GetConfig?.();
  return c ? goToAppConfig(c) : null;
};

export const wailsSaveConfig = async (c: AppConfig): Promise<AppConfig | null> => {
  if (!isWails()) return null;
  const r = await getApp().SaveConfig?.(c);
  return r ? goToAppConfig(r) : c;
};

// hover 卡片里的按钮 → 打开详细面板指定页
export const wailsOpenDashboard = async (route: string): Promise<void> => {
  if (!isWails()) return;
  await getApp().OpenDashboard?.(route);
};

// ---- 事件 ----

export const wailsOnSnapshotsUpdate = (cb: (snapshots: Snapshot[]) => void): (() => void) => {
  if (!isWails()) return () => {};
  return window.runtime!.EventsOn('snapshots:update', (data) => cb(data ?? []));
};

/** 后端窗口模式切换：'hover' | 'overview' | 'settings' | 'hidden' */
export const wailsOnUIMode = (cb: (mode: string) => void): (() => void) => {
  if (!isWails()) return () => {};
  return window.runtime!.EventsOn('ui:mode', (data) => cb(String(data ?? '')));
};

// ---- mock（仅在浏览器开发时使用） ----
function mockSnapshots(): Snapshot[] {
  const now = Date.now();
  const reset5h = Math.floor((now + 1.5 * 3600 * 1000) / 1000);
  const reset7d = Math.floor((now + 6.5 * 86400 * 1000) / 1000);
  return [
    {
      id: 'codex',
      name: 'Codex',
      shortName: 'CODEX',
      color: '#3FA8C7',
      emoji: '◉',
      primary: { usedPct: 27, remainingPct: 73, windowMinutes: 300, resetsAt: reset5h },
      secondary: { usedPct: 40, remainingPct: 60, windowMinutes: 10080, resetsAt: reset7d },
      token: {
        today: 210_000,
        near7d: 19_700_000,
        total: 505_000_000,
        currentTask: 7_140_000,
        lastTurn: 139_797,
        cacheHitRate: 97.2,
      },
      credits: { hasCredits: false, unlimited: false, balance: '0' },
      status: '执行中',
      source: 'fresh',
      ts: now,
      lastRequestAt: now,
      error: '',
    },
  ];
}

export const wailsOpenOfficialPanel = async (): Promise<void> => {
 if (isWails()) await getApp().OpenOfficialPanel?.();
};

export const wailsGetUIMode = async () => isWails() ? await getApp().GetUIMode?.() : null;
export const wailsConfirmUIMode = async (route: string) => { if (isWails()) await getApp().ConfirmUIMode?.(route); };
