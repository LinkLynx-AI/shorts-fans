"use client";

import * as Dialog from "@radix-ui/react-dialog";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

import { useHasViewerSession } from "@/entities/viewer";
import { cn } from "@/shared/lib";
import { Button, SurfacePanel } from "@/shared/ui";

import { getFanNavigationItems, resolveActiveFanNavigation } from "../model/fan-navigation";

type TemporaryProfileAuthDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

function TemporaryProfileAuthDialog({
  open,
  onOpenChange,
}: TemporaryProfileAuthDialogProps) {
  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-[#061521]/36 backdrop-blur-[2px]" />
        <Dialog.Content className="fixed inset-x-4 top-1/2 z-50 mx-auto w-full max-w-[376px] -translate-y-1/2">
          <SurfacePanel className="w-full px-5 py-6 text-foreground">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">fan access</p>
            <Dialog.Title className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
              続けるにはログインが必要です
            </Dialog.Title>
            <Dialog.Description className="mt-3 text-sm leading-6 text-muted">
              fan profile は protected surface なので、fan session を開始してから開けるようにしています。
            </Dialog.Description>

            <div className="mt-5 grid gap-2.5">
              <Button className="w-full" disabled type="button">
                サインイン / 新規登録
              </Button>
              <Dialog.Close asChild>
                <Button className="w-full" variant="secondary">
                  閉じる
                </Button>
              </Dialog.Close>
            </div>

            <p className="mt-3 text-xs leading-5 text-muted">
              TODO: 共通 auth modal manager が接続されたら、この一時 dialog を削除してそちらへ置き換える。
            </p>
          </SurfacePanel>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

/**
 * fan mode 共通の bottom navigation を表示する。
 */
export function FanBottomNavigation() {
  const hasViewerSession = useHasViewerSession();
  const [isProfileAuthDialogOpen, setIsProfileAuthDialogOpen] = useState(false);
  const pathname = usePathname();
  const activeKey = resolveActiveFanNavigation(pathname);

  return (
    <>
      <nav
        aria-label="Primary"
        className="grid grid-cols-3 border-t border-border/90 bg-tabbar-surface px-2.5 pb-[calc(3px+env(safe-area-inset-bottom,0px))] pt-2 shadow-[0_-12px_28px_rgba(36,94,132,0.08)] backdrop-blur-xl"
      >
        {getFanNavigationItems().map((item) => {
          const isActive = item.key === activeKey;

          return (
            <Link
              key={item.key}
              aria-label={item.ariaLabel}
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "grid min-h-10 place-items-center px-2 py-2 transition",
                isActive ? "text-accent-strong" : "text-accent-strong/72 hover:text-accent-strong/84",
              )}
              href={item.href}
              onClick={(event) => {
                if (item.key === "fan" && !hasViewerSession) {
                  event.preventDefault();
                  setIsProfileAuthDialogOpen(true);
                }
              }}
            >
              <item.icon aria-hidden="true" className="size-[18px]" strokeWidth={1.9} />
              <span className="sr-only">{item.ariaLabel}</span>
            </Link>
          );
        })}
      </nav>
      <TemporaryProfileAuthDialog
        onOpenChange={setIsProfileAuthDialogOpen}
        open={isProfileAuthDialogOpen}
      />
    </>
  );
}
