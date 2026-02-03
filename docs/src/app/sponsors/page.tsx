import { SponsorSection } from '@/components/sponsors'
import { sponsors, ads, githubSponsorsConfig } from '@/lib/sponsors'
import Link from 'next/link'
import Image from 'next/image'
import { Heart, Mail, ExternalLink } from 'lucide-react'
import { Metadata } from 'next'

export const metadata: Metadata = {
    title: 'Sponsors & Partners',
    description: 'Support Tharos development and explore partnership opportunities',
}

export default function SponsorsPage() {
    return (
        <div className="min-h-screen bg-gradient-to-b from-fd-background to-fd-secondary/10">
            {/* Hero Section */}
            <div className="border-b border-fd-border bg-fd-card/50 backdrop-blur">
                <div className="container mx-auto px-4 py-16 max-w-6xl">
                    <div className="text-center space-y-6">
                        <div className="inline-flex items-center gap-3 px-4 py-2 rounded-full bg-red-500/10 border border-red-500/20">
                            <Heart className="h-4 w-4 text-red-500 fill-red-500" />
                            <span className="text-sm font-medium text-red-500">Support Open Source</span>
                        </div>

                        <h1 className="text-4xl md:text-5xl font-bold tracking-tight">
                            Sponsors & Partners
                        </h1>

                        <p className="text-lg text-fd-muted-foreground max-w-2xl mx-auto">
                            Tharos is made possible by the generous support of our sponsors and partners.
                            Join us in building better security tools for developers.
                        </p>

                        {githubSponsorsConfig.enabled && (
                            <div className="pt-4">
                                <Link
                                    href={githubSponsorsConfig.url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="inline-flex items-center gap-2 px-6 py-3 rounded-lg bg-fd-primary text-fd-primary-foreground font-medium hover:bg-fd-primary/90 transition-colors"
                                >
                                    <Heart className="h-4 w-4 fill-current" />
                                    Become a Sponsor
                                    <ExternalLink className="h-4 w-4" />
                                </Link>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            <div className="container mx-auto px-4 py-16 max-w-6xl">
                {/* Current Sponsors */}
                {sponsors.length > 0 && (
                    <section className="mb-16">
                        <h2 className="text-3xl font-bold mb-8 text-center">Our Sponsors</h2>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            {sponsors.map((sponsor) => (
                                <Link
                                    key={sponsor.id}
                                    href={sponsor.url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="group relative block overflow-hidden rounded-xl border border-fd-border bg-fd-card p-6 transition-all hover:shadow-xl hover:border-fd-primary/50"
                                >
                                    <div className="flex flex-col items-center text-center space-y-4">
                                        {sponsor.logo && (
                                            <div className="relative h-20 w-20 overflow-hidden rounded-lg">
                                                <Image
                                                    src={sponsor.logo}
                                                    alt={sponsor.name}
                                                    fill
                                                    className="object-contain"
                                                />
                                            </div>
                                        )}

                                        <div>
                                            <h3 className="font-semibold text-lg group-hover:text-fd-primary transition-colors flex items-center gap-2 justify-center">
                                                {sponsor.name}
                                                <ExternalLink className="h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity" />
                                            </h3>
                                            <p className="text-sm text-fd-muted-foreground mt-2">
                                                {sponsor.description}
                                            </p>
                                        </div>
                                    </div>

                                    {/* Tier Badge */}
                                    <div className={`absolute top-4 right-4 px-2 py-1 rounded text-xs font-medium ${sponsor.tier === 'platinum' ? 'bg-purple-500/10 text-purple-500' :
                                            sponsor.tier === 'gold' ? 'bg-yellow-500/10 text-yellow-600' :
                                                sponsor.tier === 'silver' ? 'bg-gray-400/10 text-gray-500' :
                                                    'bg-orange-700/10 text-orange-700'
                                        }`}>
                                        {sponsor.tier.charAt(0).toUpperCase() + sponsor.tier.slice(1)}
                                    </div>
                                </Link>
                            ))}
                        </div>
                    </section>
                )}

                {sponsors.length === 0 && (
                    <section className="mb-16 text-center py-12">
                        <div className="max-w-md mx-auto space-y-4">
                            <Heart className="h-12 w-12 mx-auto text-fd-muted-foreground" />
                            <h3 className="text-xl font-semibold text-fd-muted-foreground">
                                Be the First Sponsor
                            </h3>
                            <p className="text-sm text-fd-muted-foreground">
                                Help support Tharos development and get your brand in front of thousands of developers.
                            </p>
                        </div>
                    </section>
                )}

                {/* Sponsorship Tiers */}
                <section className="mb-16">
                    <h2 className="text-3xl font-bold mb-8 text-center">Sponsorship Tiers</h2>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        {[
                            { tier: 'Bronze', price: '$10', color: 'orange', benefits: ['Name in README', 'Community recognition'] },
                            { tier: 'Silver', price: '$50', color: 'gray', benefits: ['Logo in docs sidebar', 'All Bronze benefits'] },
                            { tier: 'Gold', price: '$200', color: 'yellow', benefits: ['Logo on homepage', 'All Silver benefits'] },
                            { tier: 'Platinum', price: '$500+', color: 'purple', benefits: ['Premium placement', 'Custom benefits', 'All Gold benefits'] },
                        ].map((tier) => (
                            <div key={tier.tier} className="rounded-xl border border-fd-border bg-fd-card p-6 space-y-4">
                                <div className="space-y-2">
                                    <div className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${tier.color === 'purple' ? 'bg-purple-500/10 text-purple-500' :
                                            tier.color === 'yellow' ? 'bg-yellow-500/10 text-yellow-600' :
                                                tier.color === 'gray' ? 'bg-gray-400/10 text-gray-500' :
                                                    'bg-orange-700/10 text-orange-700'
                                        }`}>
                                        {tier.tier}
                                    </div>
                                    <div className="text-2xl font-bold">{tier.price}<span className="text-sm text-fd-muted-foreground font-normal">/mo</span></div>
                                </div>
                                <ul className="space-y-2 text-sm">
                                    {tier.benefits.map((benefit, i) => (
                                        <li key={i} className="flex items-start gap-2">
                                            <span className="text-green-500 mt-0.5">✓</span>
                                            <span className="text-fd-muted-foreground">{benefit}</span>
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        ))}
                    </div>
                </section>

                {/* Advertising Section */}
                <section className="mb-16">
                    <div className="rounded-2xl border border-fd-border bg-fd-card/50 p-8 md:p-12">
                        <div className="max-w-3xl mx-auto text-center space-y-6">
                            <h2 className="text-3xl font-bold">Advertising Opportunities</h2>
                            <p className="text-lg text-fd-muted-foreground">
                                Reach thousands of security-conscious developers with targeted advertising
                                in our documentation and website.
                            </p>

                            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-8">
                                <div className="space-y-2">
                                    <div className="text-3xl font-bold text-fd-primary">50K+</div>
                                    <div className="text-sm text-fd-muted-foreground">Monthly Visitors</div>
                                </div>
                                <div className="space-y-2">
                                    <div className="text-3xl font-bold text-fd-primary">10K+</div>
                                    <div className="text-sm text-fd-muted-foreground">Active Users</div>
                                </div>
                                <div className="space-y-2">
                                    <div className="text-3xl font-bold text-fd-primary">95%</div>
                                    <div className="text-sm text-fd-muted-foreground">Developer Audience</div>
                                </div>
                            </div>

                            <div className="pt-6">
                                <a
                                    href="mailto:chinonsoneft@gmail.com?subject=Tharos%20Advertising%20Inquiry"
                                    className="inline-flex items-center gap-2 px-6 py-3 rounded-lg border border-fd-border bg-fd-background font-medium hover:bg-fd-secondary transition-colors"
                                >
                                    <Mail className="h-4 w-4" />
                                    Get in Touch
                                </a>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Current Ads */}
                {ads.length > 0 && (
                    <section>
                        <h2 className="text-3xl font-bold mb-8 text-center">Featured Partners</h2>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {ads.map((ad) => (
                                <Link
                                    key={ad.id}
                                    href={ad.url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="group relative block overflow-hidden rounded-xl border border-fd-border bg-fd-card p-6 transition-all hover:shadow-lg"
                                >
                                    <div className="absolute top-3 right-3 text-[10px] uppercase tracking-wider text-fd-muted-foreground">
                                        Sponsored
                                    </div>

                                    <div className="flex items-center gap-4">
                                        {ad.logo && (
                                            <div className="relative h-16 w-16 flex-shrink-0 overflow-hidden rounded-lg">
                                                <Image
                                                    src={ad.logo}
                                                    alt={ad.name}
                                                    fill
                                                    className="object-cover"
                                                />
                                            </div>
                                        )}

                                        <div className="flex-1">
                                            <h3 className="font-semibold text-lg group-hover:text-fd-primary transition-colors">
                                                {ad.name}
                                            </h3>
                                            <p className="text-sm text-fd-muted-foreground mt-1">
                                                {ad.description}
                                            </p>
                                        </div>

                                        <ExternalLink className="h-4 w-4 text-fd-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                                    </div>
                                </Link>
                            ))}
                        </div>
                    </section>
                )}
            </div>
        </div>
    )
}
