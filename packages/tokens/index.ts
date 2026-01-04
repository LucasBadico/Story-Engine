/**
 * Design Tokens - TypeScript exports
 * For programmatic access to token values
 */

export const spacing = {
  xs: 'var(--se-space-xs)',
  sm: 'var(--se-space-sm)',
  md: 'var(--se-space-md)',
  lg: 'var(--se-space-lg)',
  xl: 'var(--se-space-xl)',
  '2xl': 'var(--se-space-2xl)',
} as const;

export const borderRadius = {
  sm: 'var(--se-radius-sm)',
  md: 'var(--se-radius-md)',
  lg: 'var(--se-radius-lg)',
} as const;

export const colors = {
  bg: 'var(--se-bg)',
  surface: 'var(--se-surface)',
  surfaceHover: 'var(--se-surface-hover)',
  text: 'var(--se-text)',
  textMuted: 'var(--se-text-muted)',
  textDisabled: 'var(--se-text-disabled)',
  border: 'var(--se-border)',
  borderHover: 'var(--se-border-hover)',
} as const;

export type Spacing = typeof spacing[keyof typeof spacing];
export type BorderRadius = typeof borderRadius[keyof typeof borderRadius];
export type Colors = typeof colors[keyof typeof colors];

