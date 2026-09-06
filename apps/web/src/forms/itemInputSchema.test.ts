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

describe('itemInputSchema length limits', () => {
  it('accepts the contract maximums', () => {
    const result = itemInputSchema.safeParse({
      name: 'a'.repeat(100),
      description: 'b'.repeat(2000),
    });
    expect(result.success).toBe(true);
  });

  it('rejects a name longer than the contract allows', () => {
    const result = itemInputSchema.safeParse({
      name: 'a'.repeat(101),
      description: '',
    });
    expect(result.success).toBe(false);
  });

  it('rejects a description longer than the contract allows', () => {
    const result = itemInputSchema.safeParse({
      name: 'Widget',
      description: 'b'.repeat(2001),
    });
    expect(result.success).toBe(false);
  });
});
