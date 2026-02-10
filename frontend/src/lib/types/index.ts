// File information from backend
export interface FileInfo {
  path: string;
  name: string;
  width: number;
  height: number;
  duration: string;
  durationSeconds: number;
  framerate: number;
  codec: string;
  audioCodec: string;
  fileSize: number;
  hasDurationMismatch: boolean;
}

// Preset definition
export interface Preset {
  name: string;
  resolution: string;
  framerate: string;
  width: number;
  height: number;
  level: string;
  fps: number;
  useSourceFps: boolean;
  useSourceRes: boolean;
}

// Encoding progress
export interface EncodingProgress {
  filename: string;
  progress: number;
  eta: string;
  currentFile: number;
  totalFiles: number;
  status: EncodingStatus;
  passNumber: number;
  totalPasses: number;
  speed: string;
}

// Encoding status
export type EncodingStatus = 'waiting' | 'encoding' | 'completed' | 'error' | 'cancelled';

// Duration mismatch check result
export interface DurationCheckResult {
  hasMismatch: boolean;
  baseDuration: string;
  tolerance: number;
  mismatchFiles: DurationMismatchInfo[];
}

export interface DurationMismatchInfo {
  path: string;
  name: string;
  duration: string;
  diff: string;
}

// App info
export interface AppInfo {
  appName: string;
  appVersion: string;
  ffmpegVersion: string;
}

// Hardware encoder information
export interface HWEncoder {
  id: string;
  name: string;
  description: string;
  available: boolean;
  priority: number;
}
