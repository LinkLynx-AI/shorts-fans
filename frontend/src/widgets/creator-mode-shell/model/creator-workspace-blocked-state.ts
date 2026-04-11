import { ApiError } from "@/shared/api";

import {
  getCreatorModeCapabilityRequiredState,
  getCreatorModeUnauthenticatedState,
  type CreatorModeShellBlockedState,
} from "./creator-mode-shell";

/**
 * creator workspace private API error から blocked state を解決する。
 */
export function resolveCreatorWorkspaceBlockedState(error: unknown): CreatorModeShellBlockedState | null {
  if (!(error instanceof ApiError) || error.code !== "http") {
    return null;
  }

  if (error.status === 401) {
    return getCreatorModeUnauthenticatedState();
  }

  if (error.status === 403) {
    return getCreatorModeCapabilityRequiredState();
  }

  return null;
}
