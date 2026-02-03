"use client";
import React, { useEffect, useState } from "react";
import { cn } from "@/lib/cn";

export const SparklesCore = (props: {
    id?: string;
    className?: string;
    background?: string;
    minSize?: number;
    maxSize?: number;
    particleDensity?: number;
    particleColor?: string;
}) => {
    const {
        id,
        className,
        background,
        minSize,
        maxSize,
        particleDensity,
        particleColor,
    } = props;
    const [init, setInit] = useState(false);

    useEffect(() => {
        setInit(true);
    }, []);

    // Simple Canvas based particle systems
    // ... using a simple effect for now to avoid tsparticles dependency issues
    // unless user specifically wants tsparticles.
    // I will write a custom lightweight canvas renderer here.

    const canvasRef = React.useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        if (!canvasRef.current) return;
        const canvas = canvasRef.current;
        const ctx = canvas.getContext("2d");
        if (!ctx) return;

        let width = (canvas.width = canvas.offsetWidth);
        let height = (canvas.height = canvas.offsetHeight);

        // Resize observer
        const handleResize = () => {
            if (canvas) {
                width = canvas.width = canvas.offsetWidth;
                height = canvas.height = canvas.offsetHeight;
            }
        };

        window.addEventListener('resize', handleResize);

        const particles: { x: number; y: number; size: number; speedX: number; speedY: number }[] = [];
        const density = particleDensity || 100;

        for (let i = 0; i < density; i++) {
            particles.push({
                x: Math.random() * width,
                y: Math.random() * height,
                size: Math.random() * ((maxSize || 3) - (minSize || 1)) + (minSize || 1),
                speedX: Math.random() * 0.5 - 0.25,
                speedY: Math.random() * 0.5 - 0.25,
            });
        }

        const animate = () => {
            ctx.clearRect(0, 0, width, height);
            particles.forEach((p) => {
                p.x += p.speedX;
                p.y += p.speedY;

                if (p.x < 0) p.x = width;
                if (p.x > width) p.x = 0;
                if (p.y < 0) p.y = height;
                if (p.y > height) p.y = 0;

                ctx.fillStyle = particleColor || "#FFFFFF";
                ctx.beginPath();
                ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
                ctx.fill();
            });
            requestAnimationFrame(animate);
        };

        const animationId = requestAnimationFrame(animate);

        return () => {
            window.removeEventListener('resize', handleResize);
            cancelAnimationFrame(animationId);
        };
    }, [maxSize, minSize, particleColor, particleDensity]);


    return (
        <canvas
            ref={canvasRef}
            id={id || "sparkles-canvas"}
            className={cn("w-full h-full", className)}
            style={{
                background: background || "transparent",
            }}
        />
    );
};
