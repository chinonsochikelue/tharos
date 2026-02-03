import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
        <div className="flex items-center gap-2 group">
          <div className="w-6 h-6 rounded bg-orange-600 flex items-center justify-center font-black italic scale-90">T</div>
          <span className="font-outfit font-black italic tracking-tighter">THAROS</span>
        </div>
      ),
    },
    links: [
      {
        text: 'Documentation',
        url: '/docs',
        active: 'nested-url',
      },
      {
        text: 'Playground',
        url: '/playground',
        active: 'nested-url',
      },
    ],
  };
}
