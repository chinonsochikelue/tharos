"use client";
import { cn } from "@/lib/cn";
import React from "react";

export const BackgroundBeams = ({ className }: { className?: string }) => {
    return (
        <div
            className={cn(
                "absolute inset-0 z-0 h-full w-full pointer-events-none overflow-hidden",
                className
            )}
        >
            <div
                className="absolute -inset-[10px] opacity-50"
                style={{
                    background: `
            repeating-linear-gradient(
              to right,
              transparent 0%,
              var(--color-slate-900) 50%,
              transparent 100%
            )
          `,
                    maskImage: `radial-gradient(ellipse at center, white, transparent)`,
                }}
            >
                <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[60rem] h-[60rem] bg-gradient-to-r from-blue-500/20 via-purple-500/20 to-cyan-500/20 rounded-full blur-[100px] animate-pulse" />
            </div>
            <div className="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_800px_at_100%_200px,#09090b,transparent)]" />
        </div>
    );
};
