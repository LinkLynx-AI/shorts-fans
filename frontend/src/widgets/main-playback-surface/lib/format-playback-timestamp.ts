export function formatPlaybackTimestamp(totalSeconds: number): string {
  const normalizedSeconds = Math.max(0, Math.floor(totalSeconds));
  const hours = Math.floor(normalizedSeconds / 3600);
  const minutes = Math.floor(normalizedSeconds / 60);
  const remainingSeconds = normalizedSeconds % 60;

  if (hours > 0) {
    return `${hours}:${String(Math.floor((normalizedSeconds % 3600) / 60)).padStart(2, "0")}:${String(remainingSeconds).padStart(2, "0")}`;
  }

  return `${String(minutes).padStart(2, "0")}:${String(remainingSeconds).padStart(2, "0")}`;
}
