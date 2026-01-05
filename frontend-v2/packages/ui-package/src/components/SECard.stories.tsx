import type { Meta, StoryObj } from '@storybook/react';
import { SECard } from './SECard';

const meta: Meta<typeof SECard> = {
  title: 'Components/SECard',
  component: SECard,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof SECard>;

export const Default: Story = {
  args: {
    children: 'Card content',
  },
};

export const WithPadding: Story = {
  args: {
    children: 'Card with padding',
    className: 'p-[var(--se-space-lg)]',
  },
};

export const Pressable: Story = {
  args: {
    children: 'Pressable card',
    isPressable: true,
    onPress: () => alert('Card pressed'),
  },
};

