<script lang="ts">
  import { outputFolder } from '../stores/settings';
  import { isEncoding } from '../stores/encoding';
  import { SelectOutputFolder, OpenOutputFolder } from '../../../wailsjs/go/main/App';

  async function handleSelectFolder() {
    try {
      const folder = await SelectOutputFolder();
      if (folder) {
        outputFolder.set(folder);
      }
    } catch (error) {
      console.error('Failed to select folder:', error);
    }
  }

  async function handleOpenFolder() {
    try {
      await OpenOutputFolder();
    } catch (error) {
      console.error('Failed to open folder:', error);
    }
  }

  function shortenPath(path: string, maxLength: number = 50): string {
    if (path.length <= maxLength) return path;
    const parts = path.split('/');
    if (parts.length <= 2) return path;

    // Keep first and last parts, abbreviate middle
    const first = parts[0] || parts[1];
    const last = parts.slice(-2).join('/');
    return `${first}/.../${last}`;
  }
</script>

<div class="output-settings">
  <label>출력 폴더</label>
  <div class="folder-row">
    <div class="folder-path" title={$outputFolder}>
      {shortenPath($outputFolder)}
    </div>
    <div class="folder-buttons">
      <button
        class="btn-secondary"
        on:click={handleSelectFolder}
        disabled={$isEncoding}
      >
        변경
      </button>
      <button
        class="btn-icon"
        on:click={handleOpenFolder}
        title="폴더 열기"
      >
        📁
      </button>
    </div>
  </div>
</div>

<style>
  .output-settings {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  label {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-secondary, #888);
  }

  .folder-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .folder-path {
    flex: 1;
    padding: 10px 12px;
    font-size: 13px;
    background: var(--bg-secondary, #1a1a1a);
    border: 1px solid var(--border-color, #444);
    border-radius: 6px;
    color: var(--text-primary, #fff);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .folder-buttons {
    display: flex;
    gap: 4px;
  }

  .btn-secondary {
    padding: 10px 16px;
    font-size: 13px;
    background: var(--bg-secondary, #1a1a1a);
    border: 1px solid var(--border-color, #444);
    border-radius: 6px;
    color: var(--text-primary, #fff);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-secondary:hover:not(:disabled) {
    background: var(--bg-hover, #222);
    border-color: var(--border-hover, #555);
  }

  .btn-secondary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-icon {
    padding: 10px;
    font-size: 16px;
    background: var(--bg-secondary, #1a1a1a);
    border: 1px solid var(--border-color, #444);
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-icon:hover {
    background: var(--bg-hover, #222);
  }
</style>
