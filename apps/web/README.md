# NodeImage Web

SvelteKit 前端，目标一比一复刻 nodeimage.com 的交互与动效。目前包含：

- 顶部导航、英雄区域、上传器、特性/流程/套餐/FAQ 模块
- TailwindCSS + Motion 动效基础结构
- 暗色模式与主题存储逻辑
- 上传器拖拽/点击队列（当前为模拟上传，用于对接 API 前的 UI 验证）

## 后续待办

- 接入真实的鉴权流程：登录、设备管理、V4 签名计算
- 根据 nodeimage.com 的实际设计微调间距、动效与图标
- 替换模拟上传逻辑为真实 API 调用，并接入批量复制功能
- 引入 Lightbox 查看器，实现滚轮缩放与拖动
- 构建独立管理后台 (`apps/admin`) 并与 API 整合

## 开发指令

```bash
npm install
npm run dev -- --open
```

## 架构说明

- `src/lib/components/Uploader.svelte`：上传器组件
- `src/lib/stores/theme.ts`：暗色模式状态
- `src/routes/+page.svelte`：首页结构

Tailwind 及 SvelteKit 配置已就绪，可直接进行样式精细化与动效复刻。
