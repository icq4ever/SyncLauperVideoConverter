<script lang="ts">
  import { onMount } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import { GetPresets, GetOutputFolder, GetFiles, CheckFFmpeg, GetAppInfo, GetAvailableEncoders } from '../wailsjs/go/main/App';

  import DropZone from './lib/components/DropZone.svelte';
  import FileList from './lib/components/FileList.svelte';
  import DurationWarning from './lib/components/DurationWarning.svelte';
  import PresetSelector from './lib/components/PresetSelector.svelte';
  import OutputSettings from './lib/components/OutputSettings.svelte';
  import ProgressBar from './lib/components/ProgressBar.svelte';
  import ControlButtons from './lib/components/ControlButtons.svelte';

  import { files, durationMismatch, clearFiles } from './lib/stores/files';
  import { setPresets, setOutputFolder, setAvailableEncoders } from './lib/stores/settings';
  import { isEncoding, updateProgress, fileCompleted, fileError, stopEncoding } from './lib/stores/encoding';

  import type { EncodingProgress, DurationCheckResult, EncodingStatus } from './lib/types';

  let ffmpegError: string | null = null;
  let appInfo: { appName: string; appVersion: string; ffmpegVersion: string } | null = null;
  let fileStatuses = new Map<string, { status: EncodingStatus; progress: number }>();

  onMount(async () => {
    // Check FFmpeg
    try {
      await CheckFFmpeg();
    } catch (error) {
      ffmpegError = 'FFmpeg를 찾을 수 없습니다. ffmpeg를 프로그램 폴더에 넣거나 PATH에 추가해주세요.';
    }

    // Load presets
    try {
      const presets = await GetPresets();
      setPresets(presets);
    } catch (error) {
      console.error('Failed to load presets:', error);
    }

    // Load available encoders
    try {
      const encoders = await GetAvailableEncoders();
      setAvailableEncoders(encoders);
    } catch (error) {
      console.error('Failed to load encoders:', error);
    }

    // Load output folder
    try {
      const folder = await GetOutputFolder();
      setOutputFolder(folder);
    } catch (error) {
      console.error('Failed to get output folder:', error);
    }

    // Get app info
    try {
      appInfo = await GetAppInfo();
    } catch (error) {
      console.error('Failed to get app info:', error);
    }

    // Set up event listeners
    EventsOn('encoding:progress', (progress: EncodingProgress) => {
      updateProgress(progress);

      // Update file status
      fileStatuses.set(progress.filename, {
        status: 'encoding',
        progress: progress.progress,
      });
      fileStatuses = fileStatuses; // trigger reactivity
    });

    EventsOn('encoding:fileComplete', (data: { success: boolean; outputPath: string; filename: string }) => {
      fileCompleted(data.filename);

      // Update file status
      fileStatuses.set(data.filename, {
        status: 'completed',
        progress: 100,
      });
      fileStatuses = fileStatuses;
    });

    EventsOn('encoding:error', (data: { error: string; filename: string }) => {
      fileError(data.filename, data.error);

      // Update file status
      fileStatuses.set(data.filename, {
        status: 'error',
        progress: 0,
      });
      fileStatuses = fileStatuses;
    });

    EventsOn('encoding:cancelled', () => {
      stopEncoding();
    });

    EventsOn('encoding:allComplete', (data: { completed: number; failed: number }) => {
      stopEncoding();
      console.log(`Encoding complete: ${data.completed} succeeded, ${data.failed} failed`);
    });

    EventsOn('duration:mismatch', (result: DurationCheckResult) => {
      durationMismatch.set(result);
    });
  });

  function handleFilesAdded(event: CustomEvent) {
    // Files are added through the store
  }

  function handleError(event: CustomEvent<{ errors?: string[]; message?: string }>) {
    const errors = event.detail.errors || [event.detail.message];
    console.error('Errors:', errors);
    // Could show a toast notification here
  }
</script>

<main class="app">
  <header class="app-header">
    <div class="header-top">
      <h1>SyncLauper VideoConverter</h1>
      <div class="header-links">
        <a href="https://synclauper.studio42.kr" target="_blank" rel="noopener">SyncLauper</a>
        <span class="link-sep">|</span>
        <a href="https://studio42.kr" target="_blank" rel="noopener">studio42</a>
      </div>
    </div>
    {#if appInfo?.ffmpegVersion}
      <div class="version">{appInfo.ffmpegVersion}</div>
    {/if}
  </header>

  {#if ffmpegError}
    <div class="error-banner">
      <span class="error-icon">❌</span>
      <span>{ffmpegError}</span>
    </div>
  {:else}
    <div class="app-content">
      <section class="section">
        <DropZone on:filesAdded={handleFilesAdded} on:error={handleError} />
      </section>

      <DurationWarning />

      <section class="section">
        <FileList {fileStatuses} />
      </section>

      <section class="section settings-section">
        <div class="settings-grid">
          <PresetSelector />
          <OutputSettings />
        </div>
      </section>

      {#if $isEncoding}
        <section class="section">
          <ProgressBar />
        </section>
      {/if}

      <section class="section controls-section">
        <ControlButtons on:error={handleError} />
      </section>
    </div>
  {/if}

  <footer class="app-footer">
    <span>SyncLauper VideoConverter v{appInfo?.appVersion || '1.0.0'}</span>
  </footer>
</main>

<style>
  :global(:root) {
    --bg-primary: #0d0d0d;
    --bg-secondary: #1a1a1a;
    --bg-tertiary: #252525;
    --bg-hover: #2a2a2a;
    --text-primary: #ffffff;
    --text-secondary: #888888;
    --border-color: #333333;
    --border-hover: #444444;
    --accent-color: #4a9eff;
    --accent-hover: #3a8eef;
    --success-color: #5cb85c;
    --warning-color: #f0ad4e;
    --warning-bg: #332b00;
    --error-color: #d9534f;
    --error-hover: #c9433f;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
  }

  :global(*) {
    box-sizing: border-box;
  }

  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    padding: 20px;
  }

  .app-header {
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .header-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .app-header h1 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .header-links {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .header-links a {
    font-size: 12px;
    color: var(--accent-color, #4a9eff);
    text-decoration: none;
    transition: color 0.2s;
  }

  .header-links a:hover {
    color: var(--accent-hover, #3a8eef);
    text-decoration: underline;
  }

  .link-sep {
    font-size: 12px;
    color: var(--text-secondary);
    opacity: 0.5;
  }

  .version {
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: 6px;
    opacity: 0.7;
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    background: rgba(217, 83, 79, 0.1);
    border: 1px solid var(--error-color);
    border-radius: 8px;
    color: var(--error-color);
    font-size: 14px;
  }

  .error-icon {
    font-size: 20px;
  }

  .app-content {
    flex: 1;
  }

  .section {
    margin-bottom: 20px;
  }

  .settings-section {
    padding: 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px solid var(--border-color);
  }

  .settings-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
  }

  @media (max-width: 600px) {
    .settings-grid {
      grid-template-columns: 1fr;
    }
  }

  .controls-section {
    padding: 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px solid var(--border-color);
  }

  .app-footer {
    margin-top: auto;
    padding-top: 16px;
    border-top: 1px solid var(--border-color);
    text-align: center;
    font-size: 12px;
    color: var(--text-secondary);
  }
</style>
