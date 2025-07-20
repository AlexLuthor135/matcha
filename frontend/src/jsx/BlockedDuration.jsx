export const getBlockedDuration = (baseTime, durationMs = null) => {
  if (!baseTime) return 'Available ✅';

  const now = new Date();
  const startTime = new Date(baseTime);
  const endTime = durationMs ? new Date(startTime.getTime() + durationMs) : startTime;

  const diffMs = endTime - now;
  if (diffMs <= 0) return 'Available ✅';

  const totalSeconds = Math.floor(diffMs / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (days > 0) return `${days}d ${hours}h ${minutes}m`;
  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
};
