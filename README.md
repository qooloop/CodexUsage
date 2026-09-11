# Codex Usage Monitor · Codex 用量监控

一款 Windows 常驻托盘工具，让你随时查看 Codex 剩余额度、重置时间和本地 Token 用量。

基于 **Go + Wails + Vue 3** 构建，使用 Windows 原生托盘与 WebView2 渲染界面。适合在使用 Codex 时放在通知区，快速了解额度消耗情况。

> 本项目为第三方工具，与 OpenAI 无隶属关系。

![Codex 用量监控总览](preview/整体预览图.png)

*界面截图使用预览数据，不代表真实账号额度。*

## 功能

- **托盘数字额度**：直接显示剩余百分比，可切换 5 小时或 7 天额度作为主指标，并可显示长期额度条。
- **悬停速览**：鼠标移到托盘图标上即可查看用量卡片，点击打开详细面板。
- **额度与重置时间**：结合本地记录和在线快照展示额度，按记录时间选择较新的数据。
- **Token 统计**：查看今日、近 7 天、累计 Token、本任务、上次对话和今日缓存命中率。
- **自动与手动刷新**：支持 5 秒、10 秒、30 秒、1 分钟、5 分钟，默认 1 分钟。
- **低额度提醒**：5 小时和 7 天额度可分别设置 10%、20%、30% 提醒阈值，也可关闭。
- **桌面集成**：支持开机启动、启动时隐藏到托盘、单实例运行和键盘快捷键。
- **代理支持**：读取 Windows 当前用户的手动系统代理，也支持代理环境变量。

<details>
<summary>查看悬停卡片、统计与设置截图</summary>

![悬停卡片](preview/dashboard-hover.png)

![用量统计](preview/dashboard-stats.png)

![设置](preview/dashboard-settings.png)

</details>

## 下载与使用

1. 在本仓库的 **Releases** 页面下载 `codex-monitor.exe`。
2. 确保 Windows 已安装 Microsoft Edge WebView2 Runtime。
3. 先使用 Codex 登录并进行一次对话，让本机生成会话记录。在线额度读取需要本地 `auth.json` 中存在 ChatGPT OAuth 登录凭证。
4. 双击运行程序。默认启动后隐藏到通知区，可在任务栏的隐藏图标区域找到它。

发布程序为便携式 EXE，无需安装本项目的 Go、Node.js 或 Wails 开发环境。配置保存在用户目录中，不随 EXE 存放。

| 操作 | 行为 |
| --- | --- |
| 悬停托盘图标 | 显示用量卡片 |
| 左键点击托盘图标 | 显示或隐藏详细面板 |
| 右键点击托盘图标 | 打开刷新、面板、设置、开机启动和退出菜单 |
| 关闭窗口 / `Escape` | 隐藏到托盘 |
| `Ctrl + R` | 刷新数据 |
| `Ctrl + D` | 打开官方用量页面 |
| 托盘菜单 → 退出 | 完全退出程序 |

键盘快捷键在应用窗口获得焦点时使用。再次启动程序会打开已有实例的面板。

## 数据来源与统计口径

### 本地记录

默认从 `%USERPROFILE%\.codex\sessions` 读取 JSONL 会话记录，启动时还会读取归档目录 `archived_sessions`。设置 `CODEX_HOME` 可指定其他 Codex 数据目录。

文件监听和每 3 秒一次的兜底扫描负责发现本地变化。**只有 Codex 将记录写入磁盘后，统计才会更新**；缩短在线刷新间隔不会让尚未写入的 Token 提前出现。

- 今日统计按本机时区的当天计算；近 7 天按最近 7 × 24 小时计算。
- 累计 Token 是本机可读取会话记录的汇总，并非账号在所有设备上的总用量。
- “本任务”和“上次对话”取自最近的可用会话事件，不跟随当前前台窗口切换；相关字段缺失时可能显示为 0。
- 缓存命中率为今日缓存输入 Token 占今日输入 Token 的比例。
- Token 卡片的进度条表示今日 Token 占近 7 天的比例，不代表官方 Token 上限。

### 在线额度与认证

程序读取 Codex 数据目录下的 `auth.json`，使用其中已有的认证信息向 ChatGPT 用量接口请求额度。遇到 HTTP 401 时会尝试刷新认证并重试一次；**刷新成功后会将新凭证写回原 `auth.json`**。应用配置文件不保存这些凭证。

在线额度依赖非公开接口，其可用性可能随服务端变化。仅有 API Key、没有上述 OAuth 凭证的环境不支持当前在线采集方式。

手动或自动刷新会先同步本地记录，再补充在线额度。网络失败时保留最近数据并显示错误状态，本地同步继续工作；自动请求失败后逐步延长重试间隔，最长 5 分钟。

剩余百分比统一向下取整，例如 **57.9% 显示为 57%**。过期记录保留原始数值与重置时间，等待新快照确认，不会自行将额度恢复为 100%。

## 设置与网络

配置文件：`%APPDATA%\CodexUsageMonitor\config.json`。

可在设置页调整刷新间隔、托盘主指标、长期额度条、悬停卡片、启动行为和通知阈值。默认关闭开机启动，开启悬停卡片，两类额度提醒阈值均为 20%。

在线请求支持 Windows 手动系统代理，包括按协议配置和绕过列表；显式设置的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY` 环境变量优先。系统代理在请求时读取，更改后无需重启程序。**暂不支持仅配置 PAC 自动代理的网络。**

## 常见问题

**启动后看不到窗口？**  
默认仅显示托盘图标，请展开任务栏隐藏图标区域并点击程序图标。

**显示无数据或 Token 为 0？**  
检查 Codex 是否已经生成会话记录，以及 `CODEX_HOME` 是否指向正确目录。部分统计需要记录中包含对应字段。

**在线刷新失败？**  
检查本地 Codex 登录状态、网络与代理设置。当前实现需要 `auth.json` 中的 OAuth 凭证；重新登录 Codex 后可手动刷新。

**为什么与官方页面的数字不完全一致？**  
本地与在线快照的采集时间可能不同，整数显示采用向下取整，Token 统计也仅覆盖本地记录。请结合更新时间与同步状态判断。

**如何更新？**  
从托盘菜单退出程序后，用新版 EXE 替换旧文件。配置保留在用户目录；目前不提供自动更新。

## 从源码构建

在 Windows 环境准备 Go（`go.mod` 声明 1.25.0）、Node.js / npm、Wails v2 CLI 和 WebView2 Runtime。Node.js 版本需满足项目所用 Vite 的要求。

在项目根目录执行：

```powershell
# 安装与项目依赖一致的 Wails CLI，并确保其位于 PATH
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0

npm install --prefix frontend

# 首次构建先生成被忽略的前端绑定与嵌入资源
wails build -s -m

# 执行完整检查并打包
./scripts/build.ps1
```

产物：`build/bin/codex-monitor.exe`。

完整构建脚本依次执行 Vue / TypeScript 类型检查、Vite 构建、Go 测试和 Wails 打包，Go 构建缓存位于 `build/cache`。修改 `frontend/src/assets/app-icon.svg` 后，使用 `./scripts/build.ps1 -RegenerateIcon` 重新生成图标并构建。

### 界面预览与回归检查

```powershell
./scripts/preview.ps1
```

保持预览运行，在另一个终端中执行：

```powershell
node frontend/node_modules/playwright/cli.js install chromium
node scripts/ui_regression.cjs
node scripts/preview_screenshots.cjs
node scripts/verify_icon.cjs
```

可通过 `PLAYWRIGHT_CHROMIUM_EXECUTABLE` 指定已有的 Chromium 路径。预览使用模拟数据，不能替代原生托盘和系统通知验收。

可选的真实账号集成检查会访问在线服务，并可能触发前述认证刷新；默认跳过，启用方式如下：

```powershell
$env:CODEX_LIVE_CHECK = '1'
try {
    go test ./internal/usage ./internal/desktop -run TestLive -v -count=1
} finally {
    Remove-Item Env:CODEX_LIVE_CHECK
}
```

## 项目结构

```text
main.go             程序入口
internal/usage/     本地记录解析、在线额度采集与统计
internal/config/    配置校验与持久化
internal/desktop/   原生托盘、窗口、通知与刷新调度
frontend/           Vue 3、TypeScript、Pinia 与界面资源
preview/               设计图、预览截图与实现核对报告
```

## 当前限制

当前实现面向 Windows，暂不提供 macOS / Linux 版本、多账号管理、云同步或自动更新。安装器尚未完成验收，当前使用便携 EXE。

多显示器、不同 DPI、任务栏位置、Explorer 重启恢复、系统通知及长期资源占用的完整桌面验收仍待完成。

反馈问题时，请附上系统版本、复现步骤和界面错误提示，并移除认证凭证与会话中的私人内容。

