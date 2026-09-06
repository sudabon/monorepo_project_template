import { describe, expect, it } from 'vitest';
import { itemInputSchema } from './itemInputSchema.ts';

describe('itemInputSchema', () => {
  it('matches the contract constraints for name and description', () => {
    expect(itemInputSchema.parse({ name: 'Widget', description: '' })).toEqual({
      name: 'Widget',
      description: '',
    });
  });

  it('rejects an empty name', () => {
    const result = itemInputSchema.safeParse({ name: '' });
    expect(result.success).toBe(false);
  });

  it('rejects NUL characters', () => {
    const result = itemInputSchema.safeParse({ name: 'a\u0000b' });
    expect(result.success).toBe(false);
  });
});
