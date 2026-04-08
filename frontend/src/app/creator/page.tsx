import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  CreatorModeShell,
  resolveCreatorModeShellState,
} from "@/widgets/creator-mode-shell";

export default async function CreatorPage() {
  const viewerState = await getFanAuthGateState();

  return <CreatorModeShell state={resolveCreatorModeShellState(viewerState.currentViewer)} />;
}
