<script lang="ts">
  import { presets, selectedPresetName, selectedPreset, availableEncoders, selectedEncoderID, selectedEncoder, qualityLevels, selectedQuality } from '../stores/settings';
  import { isEncoding } from '../stores/encoding';

  function handlePresetChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedPresetName.set(target.value);
  }

  function handleEncoderChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedEncoderID.set(target.value);
  }

  function handleQualityChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedQuality.set(Number(target.value));
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
              {encoder.name}
            </option>
          {/each}
        </select>
      </div>
    {/if}

    {#if $qualityLevels.length > 0}
      <div class="selector-group quality-group">
        <label for="quality-select">품질</label>
        <select
          id="quality-select"
          value={$selectedQuality}
          on:change={handleQualityChange}
          disabled={$isEncoding}
        >
          {#each $qualityLevels as level}
            <option value={level.value}>{level.label}</option>
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

  .selector-group.quality-group {
    flex: 0 0 auto;
    min-width: 100px;
  }

  label {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-secondary, #888);
  }

  select {
    -webkit-appearance: none;
    appearance: none;
    padding: 10px 32px 10px 12px;
    font-size: 14px;
    background: var(--bg-secondary, #1a1a1a);
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='8' viewBox='0 0 12 8'%3E%3Cpath fill='%23888' d='M1 1l5 5 5-5'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 12px center;
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
