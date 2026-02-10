<script lang="ts">
  import { currentProgress, completedCount, errorCount, errorFiles } from '../stores/encoding';
  import { totalCount } from '../stores/files';

  $: overallProgress = $totalCount > 0
    ? (($completedCount + ($currentProgress?.progress || 0) / 100) / $totalCount) * 100
    : 0;
</script>

<div class="progress-container">
  {#if $currentProgress}
    <div class="progress-header">
      <div class="file-info">
        <span class="current-file">{$currentProgress.filename}</span>
        <span class="file-count">
          ({$currentProgress.currentFile}/{$currentProgress.totalFiles})
        </span>
      </div>
      <div class="progress-stats">
        {#if $currentProgress.speed}
          <span class="speed">{$currentProgress.speed}</span>
        {/if}
        {#if $currentProgress.eta}
          <span class="eta">남은 시간: {$currentProgress.eta}</span>
        {/if}
      </div>
    </div>

    <!-- Current file progress -->
    <div class="progress-row">
      <span class="label">현재 파일</span>
      <div class="progress-bar">
        <div
          class="progress-fill"
          style="width: {$currentProgress.progress}%"
        ></div>
      </div>
      <span class="percentage">{$currentProgress.progress.toFixed(1)}%</span>
    </div>

    <!-- Pass info if multi-pass -->
    {#if $currentProgress.totalPasses > 1}
      <div class="pass-info">
        패스 {$currentProgress.passNumber}/{$currentProgress.totalPasses}
      </div>
    {/if}

    <!-- Overall progress -->
    <div class="progress-row overall">
      <span class="label">전체 진행</span>
      <div class="progress-bar">
        <div
          class="progress-fill overall-fill"
          style="width: {overallProgress}%"
        ></div>
      </div>
      <span class="percentage">{overallProgress.toFixed(1)}%</span>
    </div>

    <!-- Completion stats -->
    <div class="completion-stats">
      <span class="completed">완료: {$completedCount}</span>
      {#if $errorCount > 0}
        <span class="errors">오류: {$errorCount}</span>
      {/if}
    </div>

  {:else if $errorFiles.length === 0}
    <div class="idle-state">
      인코딩 대기 중...
    </div>
  {/if}

  <!-- Error details (always visible regardless of progress state) -->
  {#if $errorFiles.length > 0}
    <div class="error-details">
      {#each $errorFiles as err}
        <div class="error-item">
          <span class="error-filename">{err.filename}</span>
          <pre class="error-message">{err.error}</pre>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .progress-container {
    padding: 16px;
    background: var(--bg-secondary, #1a1a1a);
    border-radius: 8px;
    border: 1px solid var(--border-color, #333);
  }

  .progress-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .file-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .current-file {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary, #fff);
    max-width: 300px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-count {
    font-size: 12px;
    color: var(--text-secondary, #888);
  }

  .progress-stats {
    display: flex;
    gap: 16px;
    font-size: 12px;
    color: var(--text-secondary, #888);
  }

  .speed {
    color: var(--accent-color, #4a9eff);
  }

  .progress-row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 8px;
  }

  .progress-row.overall {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--border-color, #333);
  }

  .label {
    font-size: 12px;
    color: var(--text-secondary, #888);
    min-width: 60px;
  }

  .progress-bar {
    flex: 1;
    height: 8px;
    background: var(--bg-tertiary, #333);
    border-radius: 4px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent-color, #4a9eff);
    transition: width 0.3s ease;
  }

  .overall-fill {
    background: var(--success-color, #5cb85c);
  }

  .percentage {
    font-size: 12px;
    color: var(--text-primary, #fff);
    min-width: 50px;
    text-align: right;
  }

  .pass-info {
    font-size: 11px;
    color: var(--text-secondary, #888);
    text-align: center;
    margin-bottom: 8px;
  }

  .completion-stats {
    display: flex;
    gap: 16px;
    font-size: 12px;
    margin-top: 8px;
  }

  .completed {
    color: var(--success-color, #5cb85c);
  }

  .errors {
    color: var(--error-color, #d9534f);
  }

  .idle-state {
    text-align: center;
    padding: 20px;
    color: var(--text-secondary, #888);
    font-size: 14px;
  }

  .error-details {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--border-color, #333);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .error-item {
    background: rgba(217, 83, 79, 0.1);
    border: 1px solid rgba(217, 83, 79, 0.3);
    border-radius: 4px;
    padding: 8px 10px;
  }

  .error-filename {
    font-size: 12px;
    font-weight: 500;
    color: var(--error-color, #d9534f);
  }

  .error-message {
    font-size: 11px;
    color: var(--text-secondary, #aaa);
    margin: 4px 0 0;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 80px;
    overflow-y: auto;
  }
</style>
