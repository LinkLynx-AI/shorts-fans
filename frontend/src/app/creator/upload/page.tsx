import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  CreatorModeShell,
  resolveCreatorModeShellState,
} from "@/widgets/creator-mode-shell";
import { CreatorUploadShell } from "@/widgets/creator-upload-shell";

export default async function CreatorUploadPage() {
  const viewerState = await getFanAuthGateState();
  const state = resolveCreatorModeShellState(viewerState.currentViewer, "upload");

  if (state.kind !== "ready") {
    return <CreatorModeShell state={state} />;
  }

  return <CreatorUploadShell />;
}
