import { writable, derived } from 'svelte/store';
import type { EncodingProgress, EncodingStatus } from '../types';

// Encoding state
export interface EncodingState {
  isEncoding: boolean;
  currentProgress: EncodingProgress | null;
  completedFiles: string[];
  errorFiles: { filename: string; error: string }[];
}

const initialState: EncodingState = {
  isEncoding: false,
  currentProgress: null,
  completedFiles: [],
  errorFiles: [],
};

export const encodingState = writable<EncodingState>(initialState);

// Derived stores
export const isEncoding = derived(encodingState, $state => $state.isEncoding);
export const currentProgress = derived(encodingState, $state => $state.currentProgress);
export const completedCount = derived(encodingState, $state => $state.completedFiles.length);
export const errorCount = derived(encodingState, $state => $state.errorFiles.length);

// Helper functions
export function startEncoding() {
  encodingState.update(state => ({
    ...state,
    isEncoding: true,
    currentProgress: null,
    completedFiles: [],
    errorFiles: [],
  }));
}

export function updateProgress(progress: EncodingProgress) {
  encodingState.update(state => ({
    ...state,
    currentProgress: progress,
  }));
}

export function fileCompleted(filename: string) {
  encodingState.update(state => ({
    ...state,
    completedFiles: [...state.completedFiles, filename],
  }));
}

export function fileError(filename: string, error: string) {
  encodingState.update(state => ({
    ...state,
    errorFiles: [...state.errorFiles, { filename, error }],
  }));
}

export function stopEncoding() {
  encodingState.update(state => ({
    ...state,
    isEncoding: false,
  }));
}

export function resetEncoding() {
  encodingState.set(initialState);
}
