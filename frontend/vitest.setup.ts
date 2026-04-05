import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { createElement, type AnchorHTMLAttributes, type ReactNode } from "react";
import { afterEach, vi } from "vitest";

type MockedLinkProps = AnchorHTMLAttributes<HTMLAnchorElement> & {
  children?: ReactNode;
  href: string;
};

const mockedUsePathname = vi.fn(() => "/");
const mockedRouter = {
  back: vi.fn(),
  push: vi.fn(),
};

afterEach(() => {
  cleanup();
  mockedRouter.back.mockReset();
  mockedRouter.push.mockReset();
});

vi.mock("next/navigation", () => ({
  usePathname: mockedUsePathname,
  useRouter: () => mockedRouter,
}));

vi.mock("next/link", () => ({
  default: ({ children, href, ...props }: MockedLinkProps) =>
    createElement(
      "a",
      {
        ...props,
        href,
      },
      children,
    ),
}));
