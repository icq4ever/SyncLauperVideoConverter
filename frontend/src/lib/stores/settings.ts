import { writable, derived } from 'svelte/store';
import type { Preset, HWEncoder } from '../types';

// Available presets
export const presets = writable<Preset[]>([]);

// Selected preset name
export const selectedPresetName = writable<string>('원본 설정 유지');

// Output folder
export const outputFolder = writable<string>('');

// Available encoders
export const availableEncoders = writable<HWEncoder[]>([]);

// Selected encoder ID
export const selectedEncoderID = writable<string>('libx265');

// Derived store for selected preset
export const selectedPreset = derived(
  [presets, selectedPresetName],
  ([$presets, $selectedPresetName]) => {
    return $presets.find(p => p.name === $selectedPresetName) || null;
  }
);

// Derived store for selected encoder
export const selectedEncoder = derived(
  [availableEncoders, selectedEncoderID],
  ([$availableEncoders, $selectedEncoderID]) => {
    return $availableEncoders.find(e => e.id === $selectedEncoderID) || null;
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

export function setAvailableEncoders(encoders: HWEncoder[]) {
  availableEncoders.set(encoders);
  // Auto-select best available hardware encoder
  if (encoders.length > 0) {
    const best = encoders.reduce((a, b) => a.priority > b.priority ? a : b);
    selectedEncoderID.set(best.id);
  }
}

export function selectEncoder(encoderID: string) {
  selectedEncoderID.set(encoderID);
}
