"use client";

import * as Dialog from "@radix-ui/react-dialog";
import type { ReactNode } from "react";

type BottomSheetDialogContentProps = {
  children: ReactNode;
  description: string;
  title: string;
};

/**
 * fan / creator surface で共通利用する bottom sheet の shell を描画する。
 */
export function BottomSheetDialogContent({
  children,
  description,
  title,
}: BottomSheetDialogContentProps) {
  return (
    <Dialog.Portal>
      <Dialog.Overlay className="fixed inset-y-0 left-1/2 z-40 w-full max-w-[408px] -translate-x-1/2 bg-[rgba(77,132,166,0.22)] backdrop-blur-[8px]" />
      <Dialog.Content className="fixed bottom-3 left-1/2 z-50 w-[calc(100vw-24px)] max-w-[384px] -translate-x-1/2 rounded-[28px] border border-[rgba(217,226,232,0.94)] bg-[rgba(255,255,255,0.98)] p-[10px_10px_14px] shadow-[0_18px_42px_rgba(6,21,33,0.12)]">
        <Dialog.Title className="sr-only">{title}</Dialog.Title>
        <Dialog.Description className="sr-only">{description}</Dialog.Description>

        <div
          aria-hidden="true"
          className="mx-auto mb-3 h-1 w-10 rounded-full bg-[rgba(6,21,33,0.16)]"
        />

        {children}
      </Dialog.Content>
    </Dialog.Portal>
  );
}
