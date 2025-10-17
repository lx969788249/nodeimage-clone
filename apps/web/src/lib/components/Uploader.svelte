<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fade } from 'svelte/transition';

  interface UploadItem {
    file: File;
    status: 'pending' | 'uploading' | 'complete' | 'error';
    progress: number;
    url?: string;
    error?: string;
  }

  const dispatch = createEventDispatcher();
  let items: UploadItem[] = [];
  let dragging = false;

  const acceptMime = 'image/png,image/jpeg,image/webp,image/gif,image/svg+xml,image/avif';

  function onDrop(event: DragEvent) {
    event.preventDefault();
    dragging = false;
    if (!event.dataTransfer?.files) return;
    queueFiles(event.dataTransfer.files);
  }

  function onInput(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files) return;
    queueFiles(input.files);
    input.value = '';
  }

  function queueFiles(fileList: FileList) {
    const queued = Array.from(fileList)
      .filter((file) => acceptMime.split(',').includes(file.type))
      .map((file) => ({ file, status: 'pending', progress: 0 }) as UploadItem);
    items = [...items, ...queued];
    if (queued.length) {
      dispatch('queued', queued.map((item) => item.file));
    }
  }

  async function simulateUpload(item: UploadItem) {
    item.status = 'uploading';
    dispatch('upload', { file: item.file });
    for (let i = 1; i <= 10; i += 1) {
      await new Promise((resolve) => setTimeout(resolve, 120));
      item.progress = i * 10;
    }
    item.status = 'complete';
    item.url = URL.createObjectURL(item.file);
    dispatch('complete', { file: item.file, url: item.url });
    items = [...items];
  }

  async function startUploads() {
    for (const item of items.filter((it) => it.status === 'pending')) {
      await simulateUpload(item);
    }
  }
</script>

<div
  class="group relative flex flex-col items-center justify-center rounded-3xl border border-dashed border-neutral-300 bg-white/60 p-10 text-center shadow-sm transition hover:border-brand-400 hover:shadow-xl dark:border-neutral-700 dark:bg-neutral-900/60"
  on:dragover|preventDefault={() => (dragging = true)}
  on:dragleave={() => (dragging = false)}
  on:drop={onDrop}
>
  <div class="flex h-16 w-16 items-center justify-center rounded-full bg-brand-500/10 text-brand-500 dark:bg-brand-400/10">
    <span class="text-2xl">⬆️</span>
  </div>
  <h3 class="mt-4 text-2xl font-semibold text-neutral-900 dark:text-white">快速上传你的图片</h3>
  <p class="mt-2 text-sm text-neutral-500 dark:text-neutral-400">支持 PNG, JPG, WebP, GIF, SVG, AVIF。单张最大 20MB。</p>
  <div class="mt-6 flex flex-wrap items-center justify-center gap-3">
    <label class="cursor-pointer rounded-full bg-brand-500 px-6 py-2 text-sm font-semibold text-white transition hover:bg-brand-600">
      选择文件
      <input class="hidden" type="file" multiple accept={acceptMime} on:change={onInput} />
    </label>
    <button
      class="rounded-full border border-brand-500 px-6 py-2 text-sm font-semibold text-brand-500 transition hover:bg-brand-50 dark:hover:bg-brand-900/30"
      on:click={startUploads}
      disabled={!items.some((item) => item.status === 'pending')}
    >
      开始上传
    </button>
  </div>
  {#if dragging}
    <div in:fade out:fade class="absolute inset-0 rounded-3xl border-2 border-brand-400 border-dashed bg-brand-400/10" />
  {/if}

  {#if items.length}
    <div class="mt-8 w-full space-y-2">
      {#each items as item (item.file.name)}
        <div class="rounded-2xl border border-neutral-200 bg-white/90 p-4 shadow-sm dark:border-neutral-800 dark:bg-neutral-900/50">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-neutral-800 dark:text-neutral-100">{item.file.name}</p>
              <p class="text-xs text-neutral-400">{(item.file.size / 1024 / 1024).toFixed(2)} MB</p>
            </div>
            <span class="text-xs uppercase text-brand-500">
              {item.status === 'pending' && '待上传'}
              {item.status === 'uploading' && `${item.progress}%`}
              {item.status === 'complete' && '完成'}
              {item.status === 'error' && '失败'}
            </span>
          </div>
          <div class="mt-3 h-2 rounded-full bg-neutral-200 dark:bg-neutral-800">
            <div
              class="h-full rounded-full bg-brand-500 transition-all"
              style={`width: ${item.status === 'complete' ? 100 : item.progress}%`}
            />
          </div>
          {#if item.url && item.file.type.startsWith('image/')}
            <img class="mt-3 h-24 w-24 rounded-xl object-cover" src={item.url} alt={item.file.name} />
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
