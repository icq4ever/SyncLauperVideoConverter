<script lang="ts">
  import { durationMismatch } from '../stores/files';

  let dismissed = false;

  function dismiss() {
    dismissed = true;
  }

  // Reset dismissed state when mismatch changes
  $: if ($durationMismatch) {
    dismissed = false;
  }
</script>

{#if $durationMismatch?.hasMismatch && !dismissed}
  <div class="duration-warning">
    <div class="warning-icon">⚠️</div>
    <div class="warning-content">
      <div class="warning-title">동영상 길이가 일치하지 않습니다</div>
      <div class="warning-desc">
        동기화 재생에 문제가 발생할 수 있습니다. 기준 길이: {$durationMismatch.baseDuration}
      </div>
      <div class="mismatch-list">
        {#each $durationMismatch.mismatchFiles as file}
          <span class="mismatch-item">
            {file.name}: {file.duration} ({file.diff})
          </span>
        {/each}
      </div>
    </div>
    <button class="dismiss-btn" on:click={dismiss} title="닫기">✕</button>
  </div>
{/if}

<style>
  .duration-warning {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
    background: var(--warning-bg, #332b00);
    border: 1px solid var(--warning-color, #f0ad4e);
    border-radius: 6px;
    margin-bottom: 16px;
  }

  .warning-icon {
    font-size: 20px;
    flex-shrink: 0;
  }

  .warning-content {
    flex: 1;
  }

  .warning-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--warning-color, #f0ad4e);
    margin-bottom: 4px;
  }

  .warning-desc {
    font-size: 12px;
    color: var(--text-secondary, #ccc);
    margin-bottom: 8px;
  }

  .mismatch-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .mismatch-item {
    font-size: 11px;
    padding: 2px 8px;
    background: var(--bg-tertiary, #222);
    border-radius: 4px;
    color: var(--text-secondary, #aaa);
  }

  .dismiss-btn {
    background: none;
    border: none;
    color: var(--text-secondary, #888);
    font-size: 14px;
    cursor: pointer;
    padding: 4px;
    opacity: 0.7;
    transition: opacity 0.2s ease;
  }

  .dismiss-btn:hover {
    opacity: 1;
  }
</style>
