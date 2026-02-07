import { writable, derived } from 'svelte/store';
import type { FileInfo, DurationCheckResult } from '../types';

// Store for video files
export const files = writable<FileInfo[]>([]);

// Store for selected files (by path)
export const selectedFiles = writable<Set<string>>(new Set());

// Store for duration mismatch warning
export const durationMismatch = writable<DurationCheckResult | null>(null);

// Derived store for selected file count
export const selectedCount = derived(
  [files, selectedFiles],
  ([$files, $selectedFiles]) => {
    return $files.filter(f => $selectedFiles.has(f.path)).length;
  }
);

// Derived store for total file count
export const totalCount = derived(files, $files => $files.length);

// Helper functions
export function addFiles(newFiles: FileInfo[]) {
  files.update(current => {
    const existingPaths = new Set(current.map(f => f.path));
    const uniqueNewFiles = newFiles.filter(f => !existingPaths.has(f.path));
    return [...current, ...uniqueNewFiles];
  });

  // Auto-select new files
  selectedFiles.update(selected => {
    newFiles.forEach(f => selected.add(f.path));
    return selected;
  });
}

export function updateFileMetadata(updatedFile: FileInfo) {
  files.update(current =>
    current.map(f =>
      f.path === updatedFile.path ? { ...f, ...updatedFile } : f
    )
  );
}

export function removeFile(path: string) {
  files.update(current => current.filter(f => f.path !== path));
  selectedFiles.update(selected => {
    selected.delete(path);
    return selected;
  });
}

export function clearFiles() {
  files.set([]);
  selectedFiles.set(new Set());
  durationMismatch.set(null);
}

export function toggleFileSelection(path: string) {
  selectedFiles.update(selected => {
    if (selected.has(path)) {
      selected.delete(path);
    } else {
      selected.add(path);
    }
    return selected;
  });
}

export function selectAll() {
  files.subscribe(current => {
    selectedFiles.set(new Set(current.map(f => f.path)));
  })();
}

export function deselectAll() {
  selectedFiles.set(new Set());
}

export function getSelectedFilePaths(): string[] {
  let paths: string[] = [];
  selectedFiles.subscribe(selected => {
    paths = Array.from(selected);
  })();
  return paths;
}
