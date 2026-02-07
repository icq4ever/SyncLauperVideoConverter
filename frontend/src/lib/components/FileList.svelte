<script lang="ts">
  import FileItem from './FileItem.svelte';
  import { files, selectedFiles, removeFile, toggleFileSelection, selectAll, deselectAll } from '../stores/files';
  import { RemoveFile } from '../../../wailsjs/go/main/App';
  import { isEncoding } from '../stores/encoding';
  import type { EncodingStatus } from '../types';

  export let fileStatuses: Map<string, { status: EncodingStatus; progress: number }> = new Map();

  function handleToggle(event: CustomEvent<{ path: string }>) {
    toggleFileSelection(event.detail.path);
  }

  async function handleRemove(event: CustomEvent<{ path: string }>) {
    const { path } = event.detail;
    try {
      await RemoveFile(path);
      removeFile(path);
    } catch (error) {
      console.error('Failed to remove file:', error);
    }
  }

  function getFileStatus(path: string): EncodingStatus {
    return fileStatuses.get(path)?.status || 'waiting';
  }

  function getFileProgress(path: string): number {
    return fileStatuses.get(path)?.progress || 0;
  }

  $: allSelected = $files.length > 0 && $files.every(f => $selectedFiles.has(f.path));
  $: someSelected = $files.some(f => $selectedFiles.has(f.path)) && !allSelected;
</script>

<div class="file-list-container">
  {#if $files.length > 0}
    <div class="file-list-header">
      <label class="select-all">
        <input
          type="checkbox"
          checked={allSelected}
          indeterminate={someSelected}
          on:change={() => allSelected ? deselectAll() : selectAll()}
          disabled={$isEncoding}
        />
        <span>전체 선택 ({$selectedFiles.size}/{$files.length})</span>
      </label>
    </div>

    <div class="file-list">
      {#each $files as file (file.path)}
        <FileItem
          {file}
          selected={$selectedFiles.has(file.path)}
          status={getFileStatus(file.path)}
          progress={getFileProgress(file.path)}
          on:toggle={handleToggle}
          on:remove={handleRemove}
        />
      {/each}
    </div>
  {:else}
    <div class="empty-state">
      <p>추가된 파일이 없습니다</p>
    </div>
  {/if}
</div>

<style>
  .file-list-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .file-list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: var(--bg-secondary, #1a1a1a);
    border-radius: 6px;
  }

  .select-all {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--text-secondary, #888);
    cursor: pointer;
  }

  .select-all input {
    width: 16px;
    height: 16px;
    accent-color: var(--accent-color, #4a9eff);
  }

  .file-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-height: 300px;
    overflow-y: auto;
    padding-right: 4px;
  }

  .file-list::-webkit-scrollbar {
    width: 6px;
  }

  .file-list::-webkit-scrollbar-track {
    background: var(--bg-secondary, #1a1a1a);
    border-radius: 3px;
  }

  .file-list::-webkit-scrollbar-thumb {
    background: var(--border-color, #444);
    border-radius: 3px;
  }

  .empty-state {
    padding: 40px;
    text-align: center;
    color: var(--text-secondary, #888);
    font-size: 14px;
  }
</style>
