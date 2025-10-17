<script lang="ts">
  import Uploader from '$lib/components/Uploader.svelte';
  import { onMount } from 'svelte';
  import { themeStore } from '$lib/stores/theme';

  const features = [
    {
      title: '严格的文件验证',
      description: '多重魔数与 MIME 校验，保证上传文件真实可靠，杜绝伪装与木马。'
    },
    {
      title: '多设备同步',
      description: '同一账号最多 10 台设备实时登录，云端会话管理一键可控。'
    },
    {
      title: '智能缩略图',
      description: '自动生成多尺寸 WebP 缩略，配合 AVIF 原图，节省 95% 带宽。'
    },
    {
      title: 'NSFW 检测调优',
      description: '双阈值多模型联合审查，减少误判的同时保障内容安全。'
    }
  ];

  const steps = [
    '拖拽或选择图片上传，系统自动检测格式与安全风险',
    '后台立即生成多端适配缩略图，并分发到全球 CDN',
    '管理后台随时审查、批量操作，API 支持一键复制链接'
  ];

  onMount(() => {
    themeStore.syncWithSystem();
  });
</script>

<div class="min-h-screen bg-gradient-to-b from-neutral-50 via-white to-neutral-100 dark:from-neutral-950 dark:via-neutral-900 dark:to-neutral-950">
  <header class="relative z-10 mx-auto flex max-w-6xl items-center justify-between px-6 py-8">
    <div class="flex items-center gap-3">
      <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-brand-500 text-white shadow-lg">NI</div>
      <div>
        <p class="text-lg font-semibold text-neutral-900 dark:text-neutral-50">NodeImage</p>
        <p class="text-xs uppercase tracking-widest text-neutral-400">极速 · 安全 · 专业图床</p>
      </div>
    </div>
    <nav class="hidden items-center gap-8 text-sm font-medium text-neutral-600 dark:text-neutral-300 md:flex">
      <a class="hover:text-brand-500" href="#features">特性</a>
      <a class="hover:text-brand-500" href="#workflow">流程</a>
      <a class="hover:text-brand-500" href="#pricing">套餐</a>
      <a class="hover:text-brand-500" href="#faq">FAQ</a>
    </nav>
    <div class="flex items-center gap-3">
      <button
        class="hidden rounded-full border border-neutral-200 px-4 py-2 text-sm font-semibold text-neutral-700 transition hover:border-brand-500 hover:text-brand-500 dark:border-neutral-700 dark:text-neutral-200 md:block"
      >
        管理后台
      </button>
      <button
        class="rounded-full bg-brand-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-brand-500/30 transition hover:shadow-xl"
      >
        立即登录
      </button>
      <button
        class="ml-2 hidden h-10 w-10 items-center justify-center rounded-full border border-neutral-200 text-lg dark:border-neutral-700 md:inline-flex"
        on:click={() => themeStore.toggle()}
      >
        🌓
      </button>
    </div>
  </header>

  <main class="relative z-10 mx-auto max-w-6xl px-6 pb-24">
    <section class="grid gap-14 lg:grid-cols-[1.1fr_0.9fr] lg:items-center">
      <div>
        <div class="inline-flex items-center rounded-full border border-brand-500/30 bg-brand-500/10 px-4 py-1 text-xs font-semibold uppercase tracking-wider text-brand-600">
          NodeImage 2.0 全新升级
        </div>
        <h1 class="mt-6 text-5xl font-semibold leading-tight text-neutral-900 dark:text-white">
          极速、安全的 <span class="text-brand-500">专业图床</span>，
          为创作者准备
        </h1>
        <p class="mt-6 max-w-xl text-lg text-neutral-600 dark:text-neutral-300">
          统一域名部署、智能鉴黄、多端同步、极速分发。NodeImage 为你提供企业级的图片上传、存储、分发能力，一次上传，终身无忧。
        </p>
        <div class="mt-8 flex flex-wrap items-center gap-4 text-sm text-neutral-500 dark:text-neutral-400">
          <div class="flex items-center gap-2 rounded-full bg-white/80 px-4 py-2 shadow-sm dark:bg-neutral-900/70">
            <span class="text-lg">⚡</span> 平均响应 <strong class="text-neutral-800 dark:text-neutral-200">46ms</strong>
          </div>
          <div class="flex items-center gap-2 rounded-full bg-white/80 px-4 py-2 shadow-sm dark:bg-neutral-900/70">
            <span class="text-lg">🛡️</span> 99.99% SLA 保障
          </div>
          <div class="flex items-center gap-2 rounded-full bg-white/80 px-4 py-2 shadow-sm dark:bg-neutral-900/70">
            <span class="text-lg">🗄️</span> 统一域名入口
          </div>
        </div>
      </div>
      <div class="relative">
        <div class="absolute -left-8 -top-8 h-24 w-24 rounded-full bg-brand-500/10 blur-3xl" />
        <Uploader on:queued on:complete />
      </div>
    </section>

    <section id="features" class="mt-24 grid gap-8 md:grid-cols-2">
      {#each features as feature}
        <article class="rounded-3xl border border-neutral-200 bg-white/90 p-8 shadow-sm transition hover:-translate-y-1 hover:shadow-xl dark:border-neutral-800 dark:bg-neutral-900/70">
          <h3 class="text-xl font-semibold text-neutral-900 dark:text-white">{feature.title}</h3>
          <p class="mt-3 text-sm leading-relaxed text-neutral-500 dark:text-neutral-400">{feature.description}</p>
        </article>
      {/each}
    </section>

    <section id="workflow" class="mt-24 grid gap-10 rounded-3xl bg-neutral-900 px-10 py-14 text-neutral-100 shadow-xl dark:bg-black/60">
      <div class="flex items-center justify-between gap-4">
        <div>
          <h2 class="text-3xl font-semibold">端到端安全流程</h2>
          <p class="mt-2 text-sm text-neutral-300">从上传、存储到分发，每个环节都具备安全校验与监控。</p>
        </div>
        <div class="hidden items-center gap-3 text-xs uppercase tracking-widest text-neutral-400 md:flex">
          <span>签名校验</span>
          <span>·</span>
          <span>集中路由</span>
          <span>·</span>
          <span>人工审核</span>
        </div>
      </div>
      <ol class="space-y-6">
        {#each steps as step, index}
          <li class="flex items-start gap-5">
            <div class="flex h-9 w-9 items-center justify-center rounded-full bg-brand-500 text-sm font-bold text-white shadow-md">{index + 1}</div>
            <p class="text-sm leading-relaxed text-neutral-200">{step}</p>
          </li>
        {/each}
      </ol>
    </section>

    <section id="pricing" class="mt-24 grid gap-8 md:grid-cols-3">
      <div class="rounded-3xl border border-neutral-200 bg-white/90 p-8 shadow-sm dark:border-neutral-800 dark:bg-neutral-900/70">
        <h3 class="text-xl font-semibold text-neutral-900 dark:text-white">个人免费</h3>
        <p class="mt-3 text-sm text-neutral-500 dark:text-neutral-400">适合轻量使用，体验 NodeImage 的全部基础功能。</p>
        <p class="mt-8 text-4xl font-bold text-neutral-900 dark:text-white">¥0<span class="text-base font-medium text-neutral-400">/月</span></p>
        <ul class="mt-6 space-y-3 text-sm text-neutral-500 dark:text-neutral-400">
          <li>10GB 存储空间</li>
          <li>每日 200 次 API 调用</li>
          <li>基础 NSFW 检测</li>
          <li>7 日日志留存</li>
        </ul>
      </div>
      <div class="rounded-3xl border-2 border-brand-500 bg-gradient-to-br from-brand-500/10 via-white to-white p-8 shadow-lg shadow-brand-500/20 dark:from-brand-500/20 dark:via-neutral-900 dark:to-neutral-900">
        <h3 class="flex items-center gap-2 text-xl font-semibold text-neutral-900 dark:text-white">
          专业版
          <span class="rounded-full bg-brand-500 px-3 py-0.5 text-xs font-semibold text-white">最受欢迎</span>
        </h3>
        <p class="mt-3 text-sm text-neutral-600 dark:text-neutral-300">高并发、高可靠，适合团队与中型业务。</p>
        <p class="mt-8 text-4xl font-bold text-brand-600 dark:text-brand-400">¥99<span class="text-base font-medium text-neutral-400">/月</span></p>
        <ul class="mt-6 space-y-3 text-sm text-neutral-600 dark:text-neutral-300">
          <li>200GB 存储空间 + 高性能对象存储</li>
          <li>签名 URL + 设备会话管理</li>
          <li>AI NSFW 多模型融合</li>
          <li>自定义域名 & CDN 加速</li>
        </ul>
        <button class="mt-8 w-full rounded-full bg-brand-500 py-3 text-sm font-semibold text-white shadow-lg shadow-brand-500/30">开始试用</button>
      </div>
      <div class="rounded-3xl border border-neutral-200 bg-white/90 p-8 shadow-sm dark:border-neutral-800 dark:bg-neutral-900/70">
        <h3 class="text-xl font-semibold text-neutral-900 dark:text-white">企业定制</h3>
        <p class="mt-3 text-sm text-neutral-500 dark:text-neutral-400">专属 SLA、私有化部署与安全合规支持。</p>
        <p class="mt-8 text-4xl font-bold text-neutral-900 dark:text-white">联系我们</p>
        <ul class="mt-6 space-y-3 text-sm text-neutral-500 dark:text-neutral-400">
          <li>多区域一主两备部署</li>
          <li>自定义审核流程与报表</li>
          <li>专属顾问与 7x24 支持</li>
          <li>完善的 API SDK 与培训</li>
        </ul>
      </div>
    </section>

    <section id="faq" class="mt-24 grid gap-6 md:grid-cols-2">
      <div class="rounded-2xl border border-neutral-200 bg-white/80 p-6 dark:border-neutral-700 dark:bg-neutral-900/60">
        <h4 class="text-lg font-semibold text-neutral-900 dark:text-white">如何保证图片不会丢失？</h4>
        <p class="mt-2 text-sm text-neutral-500 dark:text-neutral-400">
          我们采用冗余对象存储与周期性健康检查，保障图片长期可用，并支持按需导出归档。
        </p>
      </div>
      <div class="rounded-2xl border border-neutral-200 bg-white/80 p-6 dark:border-neutral-700 dark:bg-neutral-900/60">
        <h4 class="text-lg font-semibold text-neutral-900 dark:text-white">支持哪些格式与动图？</h4>
        <p class="mt-2 text-sm text-neutral-500 dark:text-neutral-400">
          NodeImage 支持 PNG、JPG、WebP、GIF、AVIF、SVG，并能智能识别动图避免被压缩成静态图。
        </p>
      </div>
      <div class="rounded-2xl border border-neutral-200 bg-white/80 p-6 dark:border-neutral-700 dark:bg-neutral-900/60">
        <h4 class="text-lg font-semibold text-neutral-900 dark:text-white">上传 API 如何鉴权？</h4>
        <p class="mt-2 text-sm text-neutral-500 dark:text-neutral-400">
          每个请求都需要携带 V4 签名头与一次性 nonce，我们提供官方 SDK 帮助你快速对接。
        </p>
      </div>
      <div class="rounded-2xl border border-neutral-200 bg-white/80 p-6 dark:border-neutral-700 dark:bg-neutral-900/60">
        <h4 class="text-lg font-semibold text-neutral-900 dark:text-white">可以自定义品牌与样式吗？</h4>
        <p class="mt-2 text-sm text-neutral-500 dark:text-neutral-400">
          企业版支持上传自定义 Logo、主题色，还可通过 API 调整上传器样式与交互。随后的版本我们会开放完整的主题系统。
        </p>
      </div>
    </section>
  </main>

  <footer class="border-t border-neutral-200 bg-white/80 py-10 text-sm text-neutral-500 dark:border-neutral-800 dark:bg-neutral-900/70 dark:text-neutral-400">
    <div class="mx-auto flex max-w-6xl flex-col gap-6 px-6 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-neutral-700 dark:text-neutral-200">© {new Date().getFullYear()} NodeImage. All rights reserved.</p>
        <div class="mt-2 flex flex-wrap gap-4 text-xs uppercase tracking-widest">
          <a class="hover:text-brand-500" href="#">使用条款</a>
          <a class="hover:text-brand-500" href="#">隐私政策</a>
          <a class="hover:text-brand-500" href="#">安全说明</a>
        </div>
      </div>
      <div class="flex items-center gap-3 text-xs text-neutral-400">
        <span class="rounded-full bg-brand-500/10 px-3 py-1 text-brand-600">版本 2.0</span>
        <span>API 文档</span>
        <span>状态监控</span>
      </div>
    </div>
  </footer>
</div>
