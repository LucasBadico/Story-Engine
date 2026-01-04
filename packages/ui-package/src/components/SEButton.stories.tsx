import type { Meta, StoryObj } from '@storybook/react';
import { SEButton } from './SEButton';

const meta: Meta<typeof SEButton> = {
  title: 'Components/SEButton',
  component: SEButton,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof SEButton>;

export const Default: Story = {
  args: {
    children: 'Button',
  },
};

export const Primary: Story = {
  args: {
    children: 'Primary Button',
    color: 'primary',
  },
};

export const Secondary: Story = {
  args: {
    children: 'Secondary Button',
    variant: 'bordered',
  },
};

export const Loading: Story = {
  args: {
    children: 'Loading',
    isLoading: true,
  },
};

export const Disabled: Story = {
  args: {
    children: 'Disabled',
    isDisabled: true,
  },
};

