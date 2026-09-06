import { describe, expect, it } from 'vitest';
import { mapServerErrors } from './mapServerErrors.ts';

const fields = new Set(['name', 'description']);

describe('mapServerErrors', () => {
  it('maps contract field identifiers onto form fields', () => {
    expect(
      mapServerErrors(
        [{ field: 'name', message: 'Name must not be empty.' }],
        fields,
      ),
    ).toEqual({
      fieldErrors: { name: 'Name must not be empty.' },
      formErrors: [],
    });
  });

  it('keeps errors that do not match a form field', () => {
    expect(
      mapServerErrors(
        [{ field: 'owner.id', message: 'unknown owner' }],
        fields,
      ),
    ).toEqual({
      fieldErrors: {},
      formErrors: ['owner.id: unknown owner'],
    });
  });
});
