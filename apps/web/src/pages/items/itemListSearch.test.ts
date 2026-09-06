import { describe, expect, it } from 'vitest';
import { itemListSearchSchema } from './itemListSearch.ts';

describe('itemListSearchSchema', () => {
  it('fills in defaults when search params are missing', () => {
    expect(itemListSearchSchema.parse({})).toEqual({ q: '', page: 1 });
  });

  it('keeps a valid query string and page number', () => {
    expect(itemListSearchSchema.parse({ q: 'widget', page: '2' })).toEqual({
      q: 'widget',
      page: 2,
    });
  });

  it('falls back to defaults when URL parameters are invalid', () => {
    expect(itemListSearchSchema.parse({ q: 12, page: 'abc' })).toEqual({
      q: '',
      page: 1,
    });
    expect(itemListSearchSchema.parse({ page: '0' })).toEqual({
      q: '',
      page: 1,
    });
    expect(itemListSearchSchema.parse({ page: '-3' })).toEqual({
      q: '',
      page: 1,
    });
    expect(itemListSearchSchema.parse({ page: '1.5' })).toEqual({
      q: '',
      page: 1,
    });
  });
});
