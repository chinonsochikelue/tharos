"use client";
import React from "react";
import { SparklesCore } from "@/components/ui/sparkles";
import { BackgroundBeams } from "@/components/ui/background-beams";
import Image from "next/image";

export function GalaxyBackground({ children }: { children: React.ReactNode }) {
    return (
        <div className="relative w-full min-h-screen bg-slate-950 flex flex-col items-center justify-start overflow-hidden">
            {/* Background Layer */}
            <div className="absolute inset-0 w-full h-full z-0">
                <Image
                    src="/cosmic-horizon.png"
                    alt="Cosmic Horizon"
                    fill
                    className="object-cover opacity-40 mix-blend-screen"
                    priority
                />
                <div className="absolute inset-0 bg-gradient-to-tr from-slate-950 via-transparent to-slate-950/80" />
            </div>

            {/* Sparkles */}
            <div className="w-full absolute inset-0 h-full z-0">
                <SparklesCore
                    id="tsparticlesfullpage"
                    background="transparent"
                    minSize={0.6}
                    maxSize={1.4}
                    particleDensity={100}
                    className="w-full h-full"
                    particleColor="#FFFFFF"
                />
            </div>

            {/* Beams */}
            <BackgroundBeams className="z-0 opacity-40" />

            {/* Content */}
            <div className="relative z-10 w-full">
                {children}
            </div>
        </div>
    );
}
