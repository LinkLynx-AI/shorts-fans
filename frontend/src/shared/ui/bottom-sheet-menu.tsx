import * as Dialog from "@radix-ui/react-dialog";
import { Slot } from "@radix-ui/react-slot";
import type {
  ComponentPropsWithoutRef,
  ReactNode,
} from "react";

import { cn } from "@/shared/lib";

type BottomSheetMenuProps = {
  children: ReactNode;
  description: string;
  title: string;
  trigger: ReactNode;
};

type BottomSheetMenuGroupProps = ComponentPropsWithoutRef<"div">;

type BottomSheetMenuActionProps = ComponentPropsWithoutRef<"button"> & {
  asChild?: boolean;
  tone?: "danger" | "default";
  withDivider?: boolean;
};

type BottomSheetMenuCloseProps = ComponentPropsWithoutRef<typeof Dialog.Close>;

/**
 * モバイル向けの共通 bottom sheet menu を表示する。
 */
export function BottomSheetMenu({
  children,
  description,
  title,
  trigger,
}: BottomSheetMenuProps) {
  return (
    <Dialog.Root>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
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
    </Dialog.Root>
  );
}

/**
 * bottom sheet menu 内の action group を表示する。
 */
export function BottomSheetMenuGroup({
  className,
  ...props
}: BottomSheetMenuGroupProps) {
  return (
    <div
      className={cn("rounded-[24px] bg-[#f3f6f8] py-1", className)}
      {...props}
    />
  );
}

/**
 * bottom sheet menu の action row を表示する。
 */
export function BottomSheetMenuAction({
  asChild = false,
  className,
  tone = "default",
  withDivider = false,
  ...props
}: BottomSheetMenuActionProps) {
  const Comp = asChild ? Slot : "button";

  return (
    <Comp
      className={cn(
        "flex min-h-[54px] w-full items-center justify-between px-[18px] text-left text-sm font-bold transition disabled:cursor-default disabled:opacity-55",
        tone === "danger"
          ? "text-[#b2394f] hover:bg-[#fff1f3]"
          : "text-foreground hover:bg-white/65",
        withDivider ? "border-t border-[rgba(167,220,249,0.24)]" : "",
        className,
      )}
      {...(!asChild ? { type: "button" as const } : {})}
      {...props}
    />
  );
}

/**
 * bottom sheet menu 項目から close action を適用する。
 */
export function BottomSheetMenuClose(props: BottomSheetMenuCloseProps) {
  return <Dialog.Close {...props} />;
}
