import { NextResponse, type NextRequest } from "next/server";
import { auth0 } from "@/lib/auth0";
import { tenancyOf } from "@/lib/session";

const publicPaths = ["/marketplace/retirements", "/track"];

const protectedPrefixes = [
  "/batches",
  "/claims",
  "/credits",
  "/dashboard",
  "/facilities",
  "/marketplace",
  "/notifications",
  "/settings",
  "/verifier",
];

const onboardingPrefix = "/onboarding";

const covers = (prefix: string, pathname: string) =>
  pathname === prefix || pathname.startsWith(`${prefix}/`);

const isPublic = (pathname: string) =>
  publicPaths.some((prefix) => covers(prefix, pathname));

const isProtected = (pathname: string) =>
  !isPublic(pathname) &&
  protectedPrefixes.some((prefix) => covers(prefix, pathname));

export async function proxy(request: NextRequest) {
  const response = await auth0.middleware(request);
  const { pathname } = request.nextUrl;

  if (pathname.startsWith("/auth/")) {
    return response;
  }

  const guarded = isProtected(pathname);
  const onboarding = covers(onboardingPrefix, pathname);

  if (!guarded && !onboarding) {
    return response;
  }

  const session = await auth0.getSession(request);

  if (!session) {
    return NextResponse.redirect(
      new URL("/auth/login", request.nextUrl.origin),
    );
  }

  const tenancy = tenancyOf(session);

  if (!tenancy) {
    return response;
  }

  if (guarded && tenancy.needsOnboarding) {
    return NextResponse.redirect(
      new URL("/onboarding/organization", request.nextUrl.origin),
    );
  }

  if (onboarding && !tenancy.needsOnboarding) {
    return NextResponse.redirect(new URL("/dashboard", request.nextUrl.origin));
  }

  return response;
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt).*)",
  ],
};
