<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { files, selectedFiles, selectedCount, durationMismatch } from '../stores/files';
  import { isEncoding } from '../stores/encoding';
  import { selectedPreset, selectedPresetName, selectedEncoderID, selectedQuality } from '../stores/settings';
  import { StartEncoding, CancelEncoding, CheckDurationMismatch, SetEncoder, SetQuality } from '../../../wailsjs/go/main/App';
  import { startEncoding, stopEncoding } from '../stores/encoding';

  const dispatch = createEventDispatcher();

  let showConfirmDialog = false;
  let showUpscaleDialog = false;
  let warningMessages: { title: string; detail: string }[] = [];

  function checkUpscaleWarnings(): { title: string; detail: string }[] {
    const preset = $selectedPreset;
    if (!preset || (preset.useSourceRes && preset.useSourceFps)) {
      return [];
    }

    const selectedPaths = $selectedFiles;
    const selectedFileList = $files.filter(f => selectedPaths.has(f.path) && f.width > 0);

    const warnings: { title: string; detail: string }[] = [];
    const upscaleFiles: string[] = [];
    const fpsUpFiles: string[] = [];

    for (const file of selectedFileList) {
      // Resolution upscale check
      if (!preset.useSourceRes && preset.width > 0 && preset.height > 0) {
        if (file.width < preset.width || file.height < preset.height) {
          upscaleFiles.push(`${file.name} (${file.width}x${file.height})`);
        }
      }

      // Framerate upscale check
      if (!preset.useSourceFps && preset.fps > 0 && file.framerate > 0) {
        if (file.framerate < preset.fps - 0.1) {
          fpsUpFiles.push(`${file.name} (${file.framerate.toFixed(2)}fps)`);
        }
      }
    }

    if (upscaleFiles.length > 0) {
      const targetRes = `${preset.width}x${preset.height}`;
      const fileList = upscaleFiles.length <= 3
        ? upscaleFiles.join(', ')
        : `${upscaleFiles.length}개 파일`;
      warnings.push({
        title: `해상도 업스케일: ${fileList} \u2192 ${targetRes}`,
        detail: '화질 향상 없이 용량만 증가합니다.',
      });
    }

    if (fpsUpFiles.length > 0) {
      const targetFps = `${preset.fps}fps`;
      const fileList = fpsUpFiles.length <= 3
        ? fpsUpFiles.join(', ')
        : `${fpsUpFiles.length}개 파일`;
      warnings.push({
        title: `프레임레이트 변환: ${fileList} \u2192 ${targetFps}`,
        detail: '프레임 복제만 발생하며 실제 부드러움은 향상되지 않습니다.',
      });
    }

    return warnings;
  }

  async function handleStart() {
    // Check for duration mismatch first
    if ($durationMismatch?.hasMismatch) {
      showConfirmDialog = true;
      return;
    }

    // Check for upscale warnings
    const warnings = checkUpscaleWarnings();
    if (warnings.length > 0) {
      warningMessages = warnings;
      showUpscaleDialog = true;
      return;
    }

    await startEncodingProcess();
  }

  async function handleDurationConfirm() {
    showConfirmDialog = false;

    // After duration confirm, also check upscale
    const warnings = checkUpscaleWarnings();
    if (warnings.length > 0) {
      warningMessages = warnings;
      showUpscaleDialog = true;
      return;
    }

    await startEncodingProcess();
  }

  async function startEncodingProcess() {
    showConfirmDialog = false;
    showUpscaleDialog = false;

    try {
      // Set the selected encoder and quality before starting
      await SetEncoder($selectedEncoderID);
      await SetQuality($selectedQuality);
      startEncoding();
      await StartEncoding($selectedPresetName);
    } catch (error) {
      stopEncoding();
      dispatch('error', { message: String(error) });
    }
  }

  async function handleCancel() {
    try {
      await CancelEncoding();
      stopEncoding();
    } catch (error) {
      console.error('Failed to cancel encoding:', error);
    }
  }

  function handleDialogCancel() {
    showConfirmDialog = false;
    showUpscaleDialog = false;
  }
</script>

<div class="control-buttons">
  {#if $isEncoding}
    <button class="btn-cancel" on:click={handleCancel}>
      취소
    </button>
  {:else}
    <button
      class="btn-start"
      on:click={handleStart}
      disabled={$selectedCount === 0}
    >
      인코딩 시작 ({$selectedCount}개 파일)
    </button>
  {/if}
</div>

{#if showConfirmDialog}
  <div class="dialog-overlay" on:click={handleDialogCancel}>
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-icon">⚠️</span>
        <h3>동영상 길이 불일치 경고</h3>
      </div>
      <div class="dialog-content">
        <p>선택한 동영상들의 길이가 서로 다릅니다.</p>
        <p>동기화 재생에 문제가 발생할 수 있습니다.</p>
        <p>그래도 인코딩을 진행하시겠습니까?</p>
      </div>
      <div class="dialog-buttons">
        <button class="btn-secondary" on:click={handleDialogCancel}>
          취소
        </button>
        <button class="btn-warning" on:click={handleDurationConfirm}>
          계속 진행
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showUpscaleDialog}
  <div class="dialog-overlay" on:click={handleDialogCancel}>
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-icon">⚠️</span>
        <h3>업스케일 경고</h3>
      </div>
      <div class="dialog-content">
        {#each warningMessages as msg}
          <div class="warning-block">
            <p class="warning-title">{msg.title}</p>
            <p class="warning-detail">{msg.detail}</p>
          </div>
        {/each}
        <p class="confirm-text">그래도 인코딩을 진행하시겠습니까?</p>
      </div>
      <div class="dialog-buttons">
        <button class="btn-secondary" on:click={handleDialogCancel}>
          취소
        </button>
        <button class="btn-warning" on:click={startEncodingProcess}>
          계속 진행
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .control-buttons {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn-start {
    padding: 12px 24px;
    font-size: 14px;
    font-weight: 600;
    background: var(--accent-color, #4a9eff);
    border: none;
    border-radius: 6px;
    color: white;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-start:hover:not(:disabled) {
    background: var(--accent-hover, #3a8eef);
  }

  .btn-start:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-cancel {
    padding: 12px 24px;
    font-size: 14px;
    font-weight: 600;
    background: var(--error-color, #d9534f);
    border: none;
    border-radius: 6px;
    color: white;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-cancel:hover {
    background: var(--error-hover, #c9433f);
  }

  /* Dialog styles */
  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: var(--bg-primary, #111);
    border: 1px solid var(--border-color, #444);
    border-radius: 12px;
    padding: 24px;
    max-width: 480px;
    width: 90%;
  }

  .dialog-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }

  .dialog-icon {
    font-size: 24px;
  }

  .dialog-header h3 {
    margin: 0;
    font-size: 18px;
    color: var(--warning-color, #f0ad4e);
  }

  .dialog-content {
    margin-bottom: 24px;
  }

  .dialog-content p {
    margin: 0 0 8px 0;
    font-size: 14px;
    color: var(--text-secondary, #aaa);
    line-height: 1.5;
  }

  .warning-block {
    padding: 12px;
    background: var(--warning-bg, #332b00);
    border-radius: 6px;
    margin-bottom: 12px;
    border-left: 3px solid var(--warning-color, #f0ad4e);
  }

  .warning-block .warning-title {
    margin: 0 0 4px 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--warning-color, #f0ad4e);
  }

  .warning-block .warning-detail {
    margin: 0;
    font-size: 12px;
    color: var(--text-secondary, #aaa);
  }

  .confirm-text {
    margin-top: 16px !important;
    color: var(--text-primary, #fff) !important;
    font-weight: 500;
  }

  .dialog-buttons {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn-secondary {
    padding: 10px 20px;
    font-size: 14px;
    background: var(--bg-secondary, #222);
    border: 1px solid var(--border-color, #444);
    border-radius: 6px;
    color: var(--text-primary, #fff);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-secondary:hover {
    background: var(--bg-hover, #333);
  }

  .btn-warning {
    padding: 10px 20px;
    font-size: 14px;
    background: var(--warning-color, #f0ad4e);
    border: none;
    border-radius: 6px;
    color: #000;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-warning:hover {
    background: #e09d3e;
  }
</style>
