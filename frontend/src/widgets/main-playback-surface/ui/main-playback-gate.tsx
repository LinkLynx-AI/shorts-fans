"use client";

import Link from "next/link";
import { useSyncExternalStore } from "react";

import { hasMockMainAccessGrant } from "@/features/unlock-entry";
import { Button, SurfacePanel } from "@/shared/ui";

import type { MainPlaybackSurface as MainPlaybackSurfaceModel } from "../model/main-playback-surface";
import { MainPlaybackSurface } from "./main-playback-surface";

type MainPlaybackGateProps = {
  fallbackHref: string;
  grantToken: string;
  surface: MainPlaybackSurfaceModel;
};

function subscribeToMockMainAccess(): () => void {
  return () => {};
}

/**
 * main playback への mock access grant を検証する。
 */
export function MainPlaybackGate({
  fallbackHref,
  grantToken,
  surface,
}: MainPlaybackGateProps) {
  const hasAccess = useSyncExternalStore(
    subscribeToMockMainAccess,
    () => hasMockMainAccessGrant(surface.main.id, grantToken),
    () => false,
  );

  if (!hasAccess) {
    return (
      <main className="flex min-h-full items-center justify-center px-6 py-12">
        <SurfacePanel className="w-full max-w-xl px-8 py-9">
          <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">
            main locked
          </p>
          <h1 className="mt-4 font-display text-3xl font-semibold tracking-[-0.05em] text-foreground">
            この main はまだ unlock されていません。
          </h1>
          <p className="mt-3 text-sm leading-7 text-muted">
            再生権限のある short から unlock を通る導線で入り直してください。
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button asChild>
              <Link href={fallbackHref}>short に戻る</Link>
            </Button>
          </div>
        </SurfacePanel>
      </main>
    );
  }

  return <MainPlaybackSurface fallbackHref={fallbackHref} surface={surface} />;
}
