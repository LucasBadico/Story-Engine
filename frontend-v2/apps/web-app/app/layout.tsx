import type { Metadata } from 'next';
import '@story-engine/tokens/tokens.css';
import '@story-engine/tokens/tokens.web.css';
import './globals.css';

export const metadata: Metadata = {
  title: 'Story Engine',
  description: 'Story Engine Web App',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="se-root se-web">{children}</body>
    </html>
  );
}

