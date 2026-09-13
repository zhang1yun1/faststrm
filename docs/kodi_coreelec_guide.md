# FastStrm Kodi (CoreELEC) 插件使用与构建指南

本插件将 **FastStrm** 包装为一个标准的 Kodi 服务插件（`service.faststrm`），专为 **CoreELEC (Linux-arm64)** 电视盒子（如 S905X3/X4、S922X 等）深度定制。

无需开启 Docker 插件，也不需要安装 Entware，单台电视盒子即可实现「115网盘扫码登录 + 自动生成 .strm 本地库 + 302 重定向 4K 直连秒开」的完整闭环。

---

## 🌟 核心优势

- 🚀 **极轻量、零运行时依赖**：Go 纯静态交叉编译（无 CGO 依赖），二进制直接内置 Web 前端与 SQLite，内存占用仅数十兆。
- 📺 **随 Kodi 开机自启**：利用 Kodi `xbmc.service` 规范，Kodi 启动即拉起服务守护进程，Kodi 退出时优雅终止。
- 🛡️ **安全隔离**：配置和数据保存在 Kodi 插件专属存储目录（`/storage/.kodi/userdata/addon_data/service.faststrm/`），不破坏 CoreELEC 的只读根系统。
- 🎬 **自闭环极速播放**：生成的 `.strm` 文件保存到 `/storage/videos/strm`，Kodi 本地直接刮削海报。播放时通过 `127.0.0.1:8090` 快速 302 重定向到 115 官方 CDN，拖拽进度条不卡顿。

---

## 📥 安装与使用步骤

### 1. 获取插件安装包
你可以通过以下任一方式获取插件 ZIP 包（形如 `service.faststrm-1.3.1-arm64.zip`）：
- **方式一（推荐）**：在 Fork 后的 GitHub 仓库 Releases 页面直接下载自动构建好的 ZIP。
- **方式二（本地打包）**：在源码根目录下执行 `./build_kodi_addon.sh arm64`，打包文件将保存在 `dist/` 目录中。

### 2. 在 CoreELEC / Kodi 中安装
1. 将下载的 `service.faststrm-xxx-arm64.zip` 文件拷贝到 U 盘或通过局域网 SMB 共享传输到 CoreELEC 的 `/storage/` 目录中。
2. 打开 CoreELEC，进入 **系统设置 -> 插件**。
3. 如果未开启未知来源，先在 **系统 -> 插件** 中勾选 **未知来源**（Unknown sources）。
4. 点击 **从 zip 文件安装**，浏览并选中你的 zip 安装包，确认安装。
5. 安装完成后，Kodi 屏幕右下角会弹出提示：
   > **FastStrm 服务已就绪**  
   > 浏览器访问: `http://<盒子IP>:8090`

### 3. 扫码登录 115 与同步设置
1. 在同一局域网内的手机或电脑浏览器中打开 `http://<盒子IP>:8090`。
   - 默认账号：`admin`
   - 默认密码：`admin`（首次登录后建议修改）
2. 在 FastStrm Web 管理控制台中：
   - 点击 **账号管理**，使用 115 手机 App 扫码登录绑定账号。
   - 点击 **目录与同步设置**，添加需要同步的 115 文件夹。
   - 将 STRM 输出路径设置为：`/storage/videos/strm`（也可以在 Kodi 插件设置中查看或自定义）。
3. 开启实时监控或点击手动同步，FastStrm 会在十几秒内将网盘影片元数据转化为本地 `.strm` 文件。

### 4. 在 Kodi 中添加本地媒体库刮削
1. 在 Kodi 主界面选择 **视频 -> 文件 -> 添加视频...**。
2. 点击 **浏览** -> 选择 **根文件系统** -> 进入 `/storage/videos/strm`。
3. 设置该目录的内容类型（电影/剧集），选择刮削器（如 The Movie Database Python）。
4. 确认后 Kodi 将自动开始刮削海报、演职员信息。刮削完成后，直接在 Kodi 电影/剧集海报墙中点击播放，即可 4K 原画秒开！

---

## ⚙️ Kodi 端管理面板

在 Kodi 的 **插件 -> 程序插件**（或所有插件）中找到 **FastStrm Service**，点击即可呼出管理菜单：
- 🌐 **Web 管理地址**：弹窗提示当前电视盒子的 IP 与控制台链接、默认账号密码。
- ⚙️ **打开插件设置**：图形化修改监听端口（默认 8090）、自定义数据存储目录、STRM 目录。
- 📄 **查看运行日志**：无需 SSH，直接在电视屏幕上查看最新的 `faststrm.log` 运行日志。
- 🔄 **重启后台服务**：遇到网络切换或配置变更时一键重启守护进程。

---

## 🛠️ 本地构建与 CI/CD

### 本地编译
支持在 macOS / Linux 上直接编译：
```bash
# 赋予执行权限
chmod +x ./build_kodi_addon.sh

# 编译适配 CoreELEC 的 ARM64 插件包 (生成到 dist/ 目录)
./build_kodi_addon.sh arm64

# 编译双架构 (ARM64 + AMD64) 通用插件包
./build_kodi_addon.sh universal
```

### GitHub Actions 自动构建与发布
项目包含 `.github/workflows/build_kodi_addon_release.yml`：
- **定时同步**：定时检测官方上游 `wabisabi926/faststrm` 的 `go` 分支，当有更新时自动同步并触发构建。
- **手动触发**：可以在 GitHub 仓库的 Actions 页面随时手动点击 **Run workflow** 触发打包。
- **自动 Release**：构建完成后自动在 GitHub Releases 中生成版本并上传 `service.faststrm-*-arm64.zip`，可直接下载到电视盒子上使用。

---

## 📡 自动发布到 Kodi 插件库服务器 (Repository Server)

如果拥有自己的 Web/插件库服务器（类似 LitePan 的 `kodi.jukuku.xyz`），只需在 GitHub 仓库的 **Settings -> Secrets and variables -> Actions** 中配置以下 Secrets：

| Secret 变量名 | 说明 | 示例值 |
|---|---|---|
| `SERVER_HOST` | 插件库服务器 IP 或域名 | `kodi.example.com` |
| `SSH_PRIVATE_KEY` | 用于部署的 SSH 私钥 | `-----BEGIN OPENSSH PRIVATE KEY...` |
| `SERVER_USER` | SSH 登录用户名（默认为 `root`） | `root` |
| `SERVER_PORT` | SSH 端口（默认为 `22`） | `22` |
| `SERVER_PATH` | 插件库在服务器上的根目录 | `/var/www/kodi` |

配置后，每次 GitHub Actions 构建成功都会自动：
1. 生成标准插件库目录结构：`zips/service.faststrm/service.faststrm-<version>.zip`、`addon.xml`、`icon.png`；
2. 通过 `rsync` 增量同步到远程服务器；
3. 在远程服务器自动执行 `scripts/build_repo.py`，重新生成全局 `addons.xml` 与 `addons.xml.md5`；
4. 用户在 CoreELEC / Kodi 中无需手动更新，Kodi 将自动收到新版本推送并无感升级！
