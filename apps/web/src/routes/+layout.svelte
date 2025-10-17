<script lang="ts">
  import '../app.pcss';
  import { onMount } from 'svelte';
  import { themeStore } from '$lib/stores/theme';

  onMount(() => {
    themeStore.syncWithSystem();
  });
</script>

<svelte:body data-theme={themeStore.theme} />

<slot />
