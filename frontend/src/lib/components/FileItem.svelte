<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { FileInfo, EncodingStatus } from '../types';

  export let file: FileInfo;
  export let selected: boolean = true;
  export let status: EncodingStatus = 'waiting';
  export let progress: number = 0;

  const dispatch = createEventDispatcher();

  function formatFileSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  }

  function formatResolution(width: number, height: number): string {
    if (width >= 3840) return `4K (${width}×${height})`;
    if (width >= 1920) return `1080p (${width}×${height})`;
    if (width >= 1280) return `720p (${width}×${height})`;
    return `${width}×${height}`;
  }

  function formatFramerate(fps: number): string {
    if (Math.abs(fps - 23.976) < 0.01) return '23.976';
    if (Math.abs(fps - 29.97) < 0.01) return '29.97';
    if (Math.abs(fps - 59.94) < 0.01) return '59.94';
    return fps.toFixed(2).replace(/\.?0+$/, '');
  }

  function handleToggle() {
    dispatch('toggle', { path: file.path });
  }

  function handleRemove() {
    dispatch('remove', { path: file.path });
  }

  $: metadataLoaded = file.width > 0 || file.codec !== '';

  const statusIcons: Record<EncodingStatus, string> = {
    waiting: '⏳',
    encoding: '⚙️',
    completed: '✅',
    error: '❌',
    cancelled: '⛔',
  };
</script>

<div class="file-item" class:selected class:encoding={status === 'encoding'}>
  <label class="checkbox-wrapper">
    <input
      type="checkbox"
      checked={selected}
      on:change={handleToggle}
      disabled={status === 'encoding'}
    />
  </label>

  <div class="file-info">
    <div class="file-name">
      {#if file.hasDurationMismatch}
        <span class="warning-icon" title="다른 파일들과 길이가 다릅니다">⚠️</span>
      {/if}
      <span class="name">{file.name}</span>
    </div>
    <div class="file-meta">
      {#if metadataLoaded}
        <span class="resolution">{formatResolution(file.width, file.height)}</span>
        <span class="separator">•</span>
        <span class="framerate">{formatFramerate(file.framerate)}fps</span>
        <span class="separator">•</span>
        <span class="duration">{file.duration}</span>
        <span class="separator">•</span>
        <span class="size">{formatFileSize(file.fileSize)}</span>
        <span class="separator">•</span>
        <span class="codec">{file.codec.toUpperCase()}</span>
        {#if file.audioCodec}
          <span class="separator">•</span>
          <span class="codec">{file.audioCodec.toUpperCase()}</span>
        {/if}
      {:else}
        <span class="size">{formatFileSize(file.fileSize)}</span>
        <span class="separator">•</span>
        <span class="loading">{file.duration || '분석 중...'}</span>
      {/if}
    </div>
  </div>

  <div class="file-status">
    {#if status === 'encoding'}
      <div class="progress-bar">
        <div class="progress-fill" style="width: {progress}%"></div>
      </div>
      <span class="progress-text">{progress.toFixed(1)}%</span>
    {:else}
      <span class="status-icon">{statusIcons[status]}</span>
    {/if}
  </div>

  <button
    class="remove-btn"
    on:click={handleRemove}
    disabled={status === 'encoding'}
    title="제거"
  >
    ✕
  </button>
</div>

<style>
  .file-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: var(--bg-secondary, #1a1a1a);
    border-radius: 6px;
    border: 1px solid var(--border-color, #333);
    transition: all 0.2s ease;
  }

  .file-item:hover {
    border-color: var(--border-hover, #444);
  }

  .file-item.selected {
    border-color: var(--accent-color, #4a9eff);
  }

  .file-item.encoding {
    border-color: var(--warning-color, #f0ad4e);
  }

  .checkbox-wrapper {
    display: flex;
    align-items: center;
  }

  .checkbox-wrapper input {
    width: 18px;
    height: 18px;
    cursor: pointer;
    accent-color: var(--accent-color, #4a9eff);
  }

  .file-info {
    flex: 1;
    min-width: 0;
  }

  .file-name {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 4px;
  }

  .warning-icon {
    font-size: 14px;
  }

  .name {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary, #fff);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--text-secondary, #888);
    flex-wrap: wrap;
  }

  .separator {
    opacity: 0.5;
  }

  .loading {
    opacity: 0.6;
    font-style: italic;
  }

  .file-status {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 80px;
    justify-content: flex-end;
  }

  .progress-bar {
    width: 60px;
    height: 6px;
    background: var(--bg-tertiary, #333);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent-color, #4a9eff);
    transition: width 0.3s ease;
  }

  .progress-text {
    font-size: 11px;
    color: var(--text-secondary, #888);
    min-width: 40px;
    text-align: right;
  }

  .status-icon {
    font-size: 16px;
  }

  .remove-btn {
    background: none;
    border: none;
    color: var(--text-secondary, #888);
    font-size: 16px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .remove-btn:hover:not(:disabled) {
    background: var(--bg-hover, #333);
    color: var(--error-color, #ff6b6b);
  }

  .remove-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
</style>
