# sb-lxf (tsbAI Stack Kernel)

## 概述

sb-lxf 是基于 **sing-box-lx** 魔改的内核仓库，引入了 **tsbAI TUN 栈**（移植自 sing-box 1.15+ 的 Go TUN stack，融合 Clash 新栈设计思路）。

### tsbAI 栈特性

| 栈类型 | 说明 | 依赖 |
|--------|------|------|
| `system` | 系统栈（宿主内核 TCP） | 无额外依赖 |
| `gvisor` | 全 gVisor 用户态栈 | `-tags with_gvisor` |
| `mixed` | TCP=gVisor / UDP=system | `-tags with_gvisor` |
| **`go`** | **纯 Go TUN 栈（新）** | **无 gVisor 依赖** |

`go` 栈是 sing-box 1.15.0 引入的新特性：
- 不依赖 gVisor，纯 Go 实现
- 内存占用显著降低
- 需要 sing-tun v0.8+（含 `stack.go` 中的 `NewGo` 实现）

### 与 sing-box-lx 的关系

- 基于 Leadaxe/sing-box-lx v1.14.x 分支
- 添加了 tsbAI TUN 栈（`go` stack 支持）
- 内核库文件名：`libbox.aar` → 版本号 `v1.15.0-tsbai.1`

### 与 sb-AI 的关系

sb-AI 项目将此内核作为 Android 子模块引用：
```bash
git submodule add https://github.com/QWERZXCVA/sb-lxf.git android/core/sb-lxf
```

Android 侧通过 `scripts/fetch-libbox.sh` 下载预编译 AAR（类似 sb-AI 当前的 LxBox 流程）。

---

## 构建

```bash
# 构建桌面版
make -f Makefile.lx lx-build

# 构建 Android AAR
make lib_install && make lib_android
```

## 配置变更（vs 上游 sing-box-lx）

### DNS 高级功能（客户端 sb-AI 支持）
- `dns.fallback` + `dns.fallback-filter`：DNS 污染检测和回退
- `dns.fakeip-filter`：fakeIP 排除域名列表
- `dns.nameserver-policy`：按域名指定不同 DNS

这些在 sing-box 1.14+ 中均已原生支持，无需内核修改。

### TUN 栈新增 `go` 选项
- 配置字段：`inbounds[].tun.stack = "go"`
- sing-box 1.15+ 默认使用 `go` 栈（不再需要 `stack` 字段时的默认行为）
- 旧版内核（1.14.x）不支持 `go` 栈，需手动选择 `mixed`/`gvisor`/`system`
