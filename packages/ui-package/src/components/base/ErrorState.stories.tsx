import type { Meta, StoryObj } from '@storybook/react';
import { ErrorState } from './ErrorState';

const meta: Meta<typeof ErrorState> = {
  title: 'Components/Base/ErrorState',
  component: ErrorState,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof ErrorState>;

export const Default: Story = {
  args: {
    message: 'Something went wrong',
  },
};

export const WithRetry: Story = {
  args: {
    message: 'Failed to load data',
    onRetry: () => alert('Retrying...'),
  },
};

export const WithTitle: Story = {
  args: {
    title: 'Connection Error',
    message: 'Unable to connect to the server',
    onRetry: () => alert('Retrying...'),
  },
};

