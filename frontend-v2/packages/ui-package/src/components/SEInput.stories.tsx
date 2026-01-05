import type { Meta, StoryObj } from '@storybook/react';
import { SEInput } from './SEInput';

const meta: Meta<typeof SEInput> = {
  title: 'Components/SEInput',
  component: SEInput,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof SEInput>;

export const Default: Story = {
  args: {
    placeholder: 'Enter text...',
  },
};

export const WithLabel: Story = {
  args: {
    label: 'Name',
    placeholder: 'Enter your name',
  },
};

export const WithValue: Story = {
  args: {
    label: 'Email',
    value: 'user@example.com',
  },
};

export const Disabled: Story = {
  args: {
    label: 'Disabled',
    value: 'Cannot edit',
    isDisabled: true,
  },
};

