import type { Preview } from '@storybook/react';
import '../styles/tokens.css';

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="se-root se-web">
        <Story />
      </div>
    ),
  ],
};

export default preview;

