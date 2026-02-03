// Sponsor and Ad data configuration
export interface Sponsor {
    id: string
    name: string
    description: string
    logo: string
    url: string
    tier: 'platinum' | 'gold' | 'silver' | 'bronze'
    type: 'sponsor' | 'ad'
}

export const sponsors: Sponsor[] = [
    // Example platinum sponsor
    // {
    //   id: 'example-sponsor',
    //   name: 'Example Company',
    //   description: 'Leading provider of developer tools',
    //   logo: '/sponsors/example.png',
    //   url: 'https://example.com',
    //   tier: 'platinum',
    //   type: 'sponsor'
    // }
]

// Ads configuration
export const ads: Sponsor[] = [
    // Example ad
    {
      id: 'collabchron',
      name: 'CollabChron',
      description: 'Multi-Author Blog AI advanced platform',
      logo: '/collabchron_logo.jpg',
      url: 'https://collabchron.com.ng',
      tier: 'gold',
      type: 'ad'
    }
]

// GitHub Sponsors integration
export const githubSponsorsConfig = {
    enabled: true,
    username: 'chinonsochikelue',
    url: 'https://github.com/sponsors/chinonsochikelue'
}
