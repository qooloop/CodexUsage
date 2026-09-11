<script setup lang="ts">
/**
 * About - 关于
 */
import { computed } from 'vue';
import { useUsageStore } from '@/stores/usage';
import StatItem from '@/components/StatItem.vue';
import CodexLogo from '@/components/CodexLogo.vue';
import { fmtClockShort } from '@/lib/format';

const store = useUsageStore();
const s = computed(() => store.primary);
</script>

<template>
  <div class="about">
    <h2 class="title">关于</h2>

    <div class="card hero">
      <div class="logo">
        <CodexLogo :size="40" />
      </div>
      <div class="hero-text">
        <div class="name">Codex 用量监控</div>
        <div class="sub">v1.0.0 · Windows · Go + Wails + Vue 3</div>
      </div>
    </div>

    <div class="card">
      <StatItem icon="time" label="构建版本" value="v1.0.0 (2026-09-09)" />
      <StatItem icon="time" label="最近刷新" :value="fmtClockShort(s?.ts)" />
      <StatItem icon="time" label="数据来源" :value="s?.name ?? '—'" />
      <StatItem icon="status" label="运行状态" :value="s?.source === 'fresh' ? '已连接' : '已离线'" :status-color="s?.source === 'fresh' ? 'good' : 'idle'" />
    </div>

    <div class="card">
      <div class="feature">
        <div class="ft-ico">⚡</div>
        <div class="ft-body">
          <div class="ft-title">常驻托盘</div>
          <div class="ft-desc">Go + Win32 Shell_NotifyIcon，零 Electron</div>
        </div>
      </div>
      <div class="feature">
        <div class="ft-ico">🔒</div>
        <div class="ft-body">
          <div class="ft-title">用量来源</div>
          <div class="ft-desc">会话 Token 来自本地记录，额度结合在线快照</div>
        </div>
      </div>
      <div class="feature">
        <div class="ft-ico">📦</div>
        <div class="ft-body">
          <div class="ft-title">轻量桌面应用</div>
          <div class="ft-desc">WebView2 渲染，体积小、启动快</div>
        </div>
      </div>
    </div>

    <div class="copyright">© 2026 Codex Monitor · Made with ❤ by wanghuan</div>
  </div>
</template>

<style scoped>
.about {
  padding: 20px 24px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-width: 560px;
  height: 100%;
  overflow: auto;
}
.title {
  font-size: var(--fs-18);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 4px;
  letter-spacing: 0.3px;
}
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 4px 4px;
  box-shadow: var(--shadow-inner);
}
.hero {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
}
.logo {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(70, 189, 244, 0.18) 0%, rgba(53, 230, 165, 0.12) 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.hero-text {
  display: flex;
  flex-direction: column;
}
.name {
  font-size: var(--fs-16);
  font-weight: 600;
  color: var(--text-primary);
}
.sub {
  font-size: var(--fs-12);
  color: var(--text-tertiary);
  margin-top: 2px;
}
.feature {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
}
.ft-ico {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 8px;
}
.ft-body {
  display: flex;
  flex-direction: column;
}
.ft-title {
  font-size: var(--fs-14);
  color: var(--text-primary);
  font-weight: 500;
}
.ft-desc {
  font-size: var(--fs-12);
  color: var(--text-tertiary);
  margin-top: 2px;
}
.copyright {
  text-align: center;
  color: var(--text-tertiary);
  font-size: var(--fs-12);
  margin-top: 8px;
}
</style>
