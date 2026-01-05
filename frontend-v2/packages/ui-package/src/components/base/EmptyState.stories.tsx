import type { Meta, StoryObj } from '@storybook/react';
import { EmptyState } from './EmptyState';
import { SEButton } from '../SEButton';

const meta: Meta<typeof EmptyState> = {
  title: 'Components/Base/EmptyState',
  component: EmptyState,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof EmptyState>;

export const Default: Story = {
  args: {
    title: 'No items found',
  },
};

export const WithDescription: Story = {
  args: {
    title: 'No stories yet',
    description: 'Create your first story to get started',
  },
};

export const WithAction: Story = {
  args: {
    title: 'No stories yet',
    description: 'Create your first story to get started',
    action: <SEButton>Create Story</SEButton>,
  },
};

