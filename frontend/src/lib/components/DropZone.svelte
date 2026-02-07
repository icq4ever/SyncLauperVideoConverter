<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { AddFiles, OpenFileDialog, LoadAllMetadata } from '../../../wailsjs/go/main/App';
  import { OnFileDrop } from '../../../wailsjs/runtime/runtime';
  import { addFiles, updateFileMetadata } from '../stores/files';
  import type { FileInfo } from '../types';

  const dispatch = createEventDispatcher();

  let isDragging = false;
  let cancelFileDrop: (() => void) | null = null;

  onMount(() => {
    // Use Wails native file drop API
    // The CSS property --wails-drop-target: drop marks this element as a drop target
    cancelFileDrop = OnFileDrop((x: number, y: number, paths: string[]) => {
      if (paths && paths.length > 0) {
        addFilesToList(paths);
      }
      isDragging = false;
    }, true);
  });

  onDestroy(() => {
    if (cancelFileDrop) {
      cancelFileDrop();
    }
  });

  async function handleClick() {
    try {
      const paths = await OpenFileDialog();
      if (paths && paths.length > 0) {
        await addFilesToList(paths);
      }
    } catch (error) {
      console.error('Failed to open file dialog:', error);
    }
  }

  async function addFilesToList(paths: string[]) {
    try {
      const result = await AddFiles(paths);

      if (result.added && result.added.length > 0) {
        addFiles(result.added);
        dispatch('filesAdded', { files: result.added });
        // Load metadata in parallel
        loadMetadata();
      }

      if (result.errors && result.errors.length > 0) {
        dispatch('error', { errors: result.errors });
      }
    } catch (error) {
      console.error('Failed to add files:', error);
      dispatch('error', { errors: [String(error)] });
    }
  }

  async function loadMetadata() {
    try {
      const updated = await LoadAllMetadata();
      if (updated && updated.length > 0) {
        for (const info of updated) {
          updateFileMetadata(info);
        }
      }
    } catch (error) {
      console.error('Failed to load metadata:', error);
    }
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    isDragging = true;
  }

  function handleDragLeave(event: DragEvent) {
    event.preventDefault();
    isDragging = false;
  }
</script>

<div
  class="drop-zone"
  class:dragging={isDragging}
  style="--wails-drop-target: drop"
  on:dragover={handleDragOver}
  on:dragleave={handleDragLeave}
  on:click={handleClick}
  role="button"
  tabindex="0"
  on:keypress={(e) => e.key === 'Enter' && handleClick()}
>
  <div class="drop-zone-content">
    <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="17 8 12 3 7 8" />
      <line x1="12" y1="3" x2="12" y2="15" />
    </svg>
    <p class="main-text">비디오 파일을 드래그하거나 클릭하여 추가</p>
    <p class="sub-text">MP4, MOV, AVI, MKV, WebM 등 지원</p>
  </div>
</div>

<style>
  .drop-zone {
    border: 2px dashed var(--border-color, #444);
    border-radius: 8px;
    padding: 32px;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s ease;
    background: var(--bg-secondary, #1a1a1a);
  }

  .drop-zone:hover,
  .drop-zone.dragging {
    border-color: var(--accent-color, #4a9eff);
    background: var(--bg-hover, #222);
  }

  .drop-zone-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .icon {
    width: 48px;
    height: 48px;
    color: var(--text-secondary, #888);
  }

  .main-text {
    font-size: 16px;
    color: var(--text-primary, #fff);
    margin: 0;
  }

  .sub-text {
    font-size: 12px;
    color: var(--text-secondary, #888);
    margin: 0;
  }
</style>
