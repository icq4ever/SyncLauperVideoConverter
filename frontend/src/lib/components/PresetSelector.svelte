<script lang="ts">
  import { presets, selectedPresetName, selectedPreset } from '../stores/settings';
  import { isEncoding } from '../stores/encoding';

  function handleChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedPresetName.set(target.value);
  }

  function getPresetDescription(preset: typeof $selectedPreset): string {
    if (!preset) return '';
    if (preset.useSourceRes && preset.useSourceFps) {
      return '원본 해상도 및 프레임레이트 유지, HEVC 인코딩';
    }
    return `${preset.resolution} @ ${preset.framerate}fps → HEVC/MKV`;
  }
</script>

<div class="preset-selector">
  <label for="preset-select">프리셋</label>
  <select
    id="preset-select"
    value={$selectedPresetName}
    on:change={handleChange}
    disabled={$isEncoding}
  >
    {#each $presets as preset}
      <option value={preset.name}>{preset.name}</option>
    {/each}
  </select>
  {#if $selectedPreset}
    <div class="preset-info">
      {getPresetDescription($selectedPreset)}
    </div>
  {/if}
</div>

<style>
  .preset-selector {
    display: flex;
    flex-direction: column;
    gap: 8px;
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

  .preset-info {
    font-size: 12px;
    color: var(--text-secondary, #888);
    padding: 8px 12px;
    background: var(--bg-tertiary, #222);
    border-radius: 4px;
  }
</style>
