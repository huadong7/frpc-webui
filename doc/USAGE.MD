● 两个平台二进制都已更新。现在默认监听 0.0.0.0（所有网络接口）。

  使用方式：

  # Linux 直接启动（默认监听所有接口）
  ./frpc-linux-amd64 --web-port 7400

  # 可选：仅监听回环地址
  ./frpc-linux-amd64 --web-port 7400 --web-addr 127.0.0.1

  # 可选：设置认证
  ./frpc-linux-amd64 --web-port 7400 --web-user admin --web-password mypassword

  # Docker 部署
  docker compose -f docker-compose.prod.yml up -d

  完整参数列表：

  ┌─────────────────┬──────────────────────┬────────────────┐
  │      参数       │        默认值        │      说明      │
  ├─────────────────┼──────────────────────┼────────────────┤
  │ --web-addr      │ 0.0.0.0              │ 监听地址       │
  ├─────────────────┼──────────────────────┼────────────────┤
  │ --web-port      │ 7400                 │ 监听端口       │
  ├─────────────────┼──────────────────────┼────────────────┤
  │ --web-user      │ (空)                 │ 基本认证用户名 │
  ├─────────────────┼──────────────────────┼────────────────┤
  │ --web-password  │ (空)                 │ 基本认证密码   │
  ├─────────────────┼──────────────────────┼────────────────┤
  │ --profiles-path │ ./frpc_profiles.json │ 配置存储路径   │
  └─────────────────┴──────────────────────┴────────────────┘

  支持的环境变量：FRPC_WEB_ADDR、FRPC_WEB_PORT、FRPC_WEB_USER、FRPC_WEB_PASSWORD、FRPC_PROFILES_PATH

## 端口冲突检测

frpc-webui 现在支持实时检测端口冲突。在配置代理时，系统会自动查询 frps 服务器上已使用的端口，并在输入框中显示已使用的端口列表。如果输入的远程端口已被占用，会显示错误警告，避免配置冲突。

此功能需要配置 frps dashboard 地址（在 profile 的 dashboard 部分）。如果未配置 dashboard 地址，端口冲突检测将被禁用。

## Docker 镜像打包

使用 `dockerfiles/Dockerfile`（非多阶段构建），先在本机编译好 Linux 二进制再打入镜像：

```bash
# 1. 构建前端
make frpc-web

# 2. 编译 Linux amd64 二进制（嵌入 web 资源）
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -tags frpc -o bin/frpc-linux-amd64 ./cmd/frpc

# 3. 构建 Docker 镜像
docker build -t frpc-manager:latest -f dockerfiles/Dockerfile .

# 4. 导出镜像（用于迁移到其他机器）
docker save frpc-manager:latest | gzip > frpc-manager-latest.tar.gz

# 5. 目标机器加载
docker load < frpc-manager-latest.tar.gz

# 6. 启动
docker run -d --name frpc-manager -p 7400:7400 -v frpc_data:/data --restart unless-stopped frpc-manager:latest
```

### 已知问题修复 (2026-06-12)

1. **编辑 Profile 后 proxies 丢失**：编辑 Profile 保存时，前端错误地将 `proxies` 和 `visitors` 设为空数组，导致该 profile 下所有代理配置丢失。已修复：编辑模式保存时会保留原有的 proxies 和 visitors。

2. **Dashboard 连接失败无提示**：创建新代理时，如果 frps dashboard 未配置或登录失败，端口冲突检测静默失效，用户无法得知原因。已修复：创建新代理时若 dashboard 查询失败，会在 Remote Port 输入框下方显示黄色警告提示，引导用户检查 profile 中的 dashboard 配置（地址、用户名、密码）。

### UI视觉更新 (2026-06-17)

1. **赛博朋克风格背景**：添加了赛博朋克风格的背景动画效果，包括网格图案、发光光球和扫描线，增强界面视觉吸引力。

2. **页面过渡动画**：实现了页面切换时的淡入淡出过渡效果，提升用户体验流畅度。

3. **卡片发光效果**：为卡片组件添加了发光脉冲动画，增强交互反馈。

4. **样式系统优化**：重构了CSS变量和动画系统，统一了设计语言。

### 性能优化 - 移除重型视觉效果 (2026-06-18)

**问题**：赛博朋克风格的背景动画导致浏览器卡顿。

**根因**：多个CSS动画同时运行（grid-scroll、float-1/2、scan、spot-drift-1~4），大量模糊滤镜（blur(80px)、blur(100px)），以及requestAnimationFrame循环跟踪鼠标移动。

**解决方案**：移除了以下重型视觉效果：
- 鼠标跟随光效（600px青紫渐变光晕）
- 彩色光斑动画（4个不同颜色的大型模糊光斑）
- 交互式文字光效
- 网格背景动画
- 扫描线动画
- 发光光球动画
- 页面过渡动画

恢复为轻量级的静态样式，确保浏览器访问流畅。