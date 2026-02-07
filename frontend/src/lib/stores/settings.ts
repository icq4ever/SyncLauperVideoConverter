import { writable, derived } from 'svelte/store';
import type { Preset } from '../types';

// Available presets
export const presets = writable<Preset[]>([]);

// Selected preset name
export const selectedPresetName = writable<string>('원본 설정 유지');

// Output folder
export const outputFolder = writable<string>('');

// Derived store for selected preset
export const selectedPreset = derived(
  [presets, selectedPresetName],
  ([$presets, $selectedPresetName]) => {
    return $presets.find(p => p.name === $selectedPresetName) || null;
  }
);

// Helper functions
export function setPresets(newPresets: Preset[]) {
  presets.set(newPresets);
}

export function selectPreset(name: string) {
  selectedPresetName.set(name);
}

export function setOutputFolder(folder: string) {
  outputFolder.set(folder);
}
