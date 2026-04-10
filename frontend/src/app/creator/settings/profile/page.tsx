import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  CreatorModeShell,
  CreatorProfileSettingsShell,
  resolveCreatorModeShellState,
} from "@/widgets/creator-mode-shell";

export default async function CreatorProfileSettingsPage() {
  const viewerState = await getFanAuthGateState();
  const state = resolveCreatorModeShellState(viewerState.currentViewer);

  if (state.kind !== "ready") {
    return <CreatorModeShell state={state} />;
  }

  return <CreatorProfileSettingsShell />;
}
