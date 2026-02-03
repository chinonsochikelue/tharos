'use client'

import Link from 'next/link'
import Image from 'next/image'
import { sponsors, ads, githubSponsorsConfig, type Sponsor } from '@/lib/sponsors'
import { Heart, ExternalLink } from 'lucide-react'
import { cn } from '@/lib/cn'

interface SponsorCardProps {
    sponsor: Sponsor
    compact?: boolean
}

function SponsorCard({ sponsor, compact = false }: SponsorCardProps) {
    const tierColors = {
        platinum: 'from-purple-500 to-pink-500',
        gold: 'from-yellow-500 to-orange-500',
        silver: 'from-gray-400 to-gray-500',
        bronze: 'from-orange-700 to-amber-700'
    }

    return (
        <Link
            href={sponsor.url}
            target="_blank"
            rel="noopener noreferrer"
            className={cn(
                'group relative block overflow-hidden rounded-lg border border-fd-border bg-fd-card p-4 transition-all hover:shadow-lg hover:border-fd-primary/50',
                compact && 'p-3'
            )}
        >
            <div className="flex items-start gap-3">
                {sponsor.logo && (
                    <div className="relative h-12 w-12 flex-shrink-0 overflow-hidden rounded-md">
                        <Image
                            src={sponsor.logo}
                            alt={sponsor.name}
                            fill
                            className="object-cover"
                        />
                    </div>
                )}
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                        <h4 className="font-semibold text-sm truncate">{sponsor.name}</h4>
                        <ExternalLink className="h-3 w-3 text-fd-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                    </div>
                    {!compact && sponsor.description && (
                        <p className="text-xs text-fd-muted-foreground mt-1 line-clamp-2">
                            {sponsor.description}
                        </p>
                    )}
                </div>
            </div>

            {/* Tier indicator */}
            <div className={cn(
                'absolute top-0 right-0 h-1 w-full bg-gradient-to-r',
                tierColors[sponsor.tier]
            )} />
        </Link>
    )
}

interface SponsorSectionProps {
    title?: string
    showGitHub?: boolean
    showSponsors?: boolean
    showAds?: boolean
    compact?: boolean
    className?: string
}

export function SponsorSection({
    title = 'Sponsors',
    showGitHub = true,
    showSponsors = true,
    showAds = true,
    compact = false,
    className
}: SponsorSectionProps) {
    const hasSponsors = sponsors.length > 0
    const hasAds = ads.length > 0
    const showContent = (showSponsors && hasSponsors) || (showAds && hasAds) || (showGitHub && githubSponsorsConfig.enabled)

    if (!showContent) return null

    return (
        <div className={cn('space-y-4', className)}>
            <div className="flex items-center gap-2">
                <Heart className="h-4 w-4 text-red-500" />
                <h3 className="text-sm font-semibold">{title}</h3>
            </div>

            <div className="space-y-3">
                {/* GitHub Sponsors CTA */}
                {showGitHub && githubSponsorsConfig.enabled && (
                    <Link
                        href={githubSponsorsConfig.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block w-full rounded-lg border-2 border-dashed border-fd-primary/30 bg-fd-primary/5 p-4 text-center transition-all hover:border-fd-primary/50 hover:bg-fd-primary/10"
                    >
                        <Heart className="mx-auto h-6 w-6 text-fd-primary mb-2" />
                        <p className="text-sm font-medium">Become a Sponsor</p>
                        <p className="text-xs text-fd-muted-foreground mt-1">
                            Support Tharos development
                        </p>
                    </Link>
                )}

                {/* Display Sponsors */}
                {showSponsors && hasSponsors && (
                    <div className="space-y-2">
                        {sponsors.map((sponsor) => (
                            <SponsorCard key={sponsor.id} sponsor={sponsor} compact={compact} />
                        ))}
                    </div>
                )}

                {/* Display Ads */}
                {showAds && hasAds && (
                    <div className="space-y-2">
                        {ads.map((ad) => (
                            <div key={ad.id}>
                                <SponsorCard sponsor={ad} compact={compact} />
                                <p className="text-[10px] text-fd-muted-foreground text-center mt-1">
                                    Advertisement
                                </p>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    )
}

// Inline ad component for placing between content
export function InlineAd({ sponsor }: { sponsor: Sponsor }) {
    return (
        <div className="my-8 mx-auto max-w-2xl">
            <div className="rounded-lg border border-fd-border bg-fd-card/50 p-6">
                <div className="flex items-center justify-between mb-3">
                    <span className="text-[10px] uppercase tracking-wider text-fd-muted-foreground">
                        Sponsored
                    </span>
                    <ExternalLink className="h-3 w-3 text-fd-muted-foreground" />
                </div>
                <Link
                    href={sponsor.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="group block"
                >
                    <div className="flex items-center gap-4">
                        {sponsor.logo && (
                            <div className="relative h-16 w-16 flex-shrink-0 overflow-hidden rounded-lg">
                                <Image
                                    src={sponsor.logo}
                                    alt={sponsor.name}
                                    fill
                                    className="object-cover"
                                />
                            </div>
                        )}
                        <div className="flex-1">
                            <h4 className="font-semibold text-lg group-hover:text-fd-primary transition-colors">
                                {sponsor.name}
                            </h4>
                            <p className="text-sm text-fd-muted-foreground mt-1">
                                {sponsor.description}
                            </p>
                        </div>
                    </div>
                </Link>
            </div>
        </div>
    )
}
