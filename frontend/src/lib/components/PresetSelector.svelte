<script lang="ts">
  import { presets, selectedPresetName, selectedPreset, availableEncoders, selectedEncoderID, selectedEncoder } from '../stores/settings';
  import { isEncoding } from '../stores/encoding';

  function handlePresetChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedPresetName.set(target.value);
  }

  function handleEncoderChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedEncoderID.set(target.value);
  }

  function getPresetDescription(preset: typeof $selectedPreset): string {
    if (!preset) return '';
    if (preset.useSourceRes && preset.useSourceFps) {
      return '원본 해상도 및 프레임레이트 유지, HEVC 인코딩';
    }
    return `${preset.resolution} @ ${preset.framerate}fps → HEVC/MKV`;
  }

  function getEncoderIcon(encoderId: string): string {
    if (encoderId === 'libx265') return '💻';
    if (encoderId.includes('videotoolbox')) return '🍎';
    if (encoderId.includes('nvenc')) return '🟢';
    if (encoderId.includes('qsv')) return '🔵';
    if (encoderId.includes('amf')) return '🔴';
    if (encoderId.includes('vaapi')) return '🐧';
    return '⚡';
  }
</script>

<div class="preset-selector">
  <div class="selector-row">
    <div class="selector-group">
      <label for="preset-select">프리셋</label>
      <select
        id="preset-select"
        value={$selectedPresetName}
        on:change={handlePresetChange}
        disabled={$isEncoding}
      >
        {#each $presets as preset}
          <option value={preset.name}>{preset.name}</option>
        {/each}
      </select>
    </div>

    {#if $availableEncoders.length > 0}
      <div class="selector-group">
        <label for="encoder-select">인코더</label>
        <select
          id="encoder-select"
          value={$selectedEncoderID}
          on:change={handleEncoderChange}
          disabled={$isEncoding}
        >
          {#each $availableEncoders as encoder}
            <option value={encoder.id}>
              {getEncoderIcon(encoder.id)} {encoder.name}
            </option>
          {/each}
        </select>
      </div>
    {/if}
  </div>

  <div class="info-row">
    {#if $selectedPreset}
      <div class="preset-info">
        {getPresetDescription($selectedPreset)}
      </div>
    {/if}
    {#if $selectedEncoder}
      <div class="encoder-info" class:hardware={$selectedEncoder.id !== 'libx265'}>
        {$selectedEncoder.description}
      </div>
    {/if}
  </div>
</div>

<style>
  .preset-selector {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .selector-row {
    display: flex;
    gap: 12px;
  }

  .selector-group {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  label {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-secondary, #888);
  }

  select {
    padding: 10px 12px;
    font-size: 14px;
    background: var(--bg-secondary, #1a1a1a);
    border: 1px solid var(--border-color, #444);
    border-radius: 6px;
    color: var(--text-primary, #fff);
    cursor: pointer;
    transition: border-color 0.2s ease;
  }

  select:hover:not(:disabled) {
    border-color: var(--border-hover, #555);
  }

  select:focus {
    outline: none;
    border-color: var(--accent-color, #4a9eff);
  }

  select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .info-row {
    display: flex;
    gap: 8px;
  }

  .preset-info,
  .encoder-info {
    flex: 1;
    font-size: 12px;
    color: var(--text-secondary, #888);
    padding: 8px 12px;
    background: var(--bg-tertiary, #222);
    border-radius: 4px;
  }

  .encoder-info.hardware {
    background: linear-gradient(135deg, #1a2a1a 0%, #222 100%);
    border-left: 2px solid #4a9;
  }
</style>
