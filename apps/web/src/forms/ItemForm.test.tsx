import { ApiError } from '@monorepo-project-template/api-client';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ItemForm } from './ItemForm.tsx';

describe('ItemForm', () => {
  it('stops submit on client-side validation errors', async () => {
    const submit = vi.fn();
    const user = userEvent.setup();
    render(<ItemForm submit={submit} />);
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByRole('alert')).toHaveTextContent(
      '名前を入力してください',
    );
    expect(submit).not.toHaveBeenCalled();
  });

  it('exposes client-side field errors as the input accessible description', async () => {
    const submit = vi.fn();
    const user = userEvent.setup();
    render(<ItemForm submit={submit} />);
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByLabelText('名前')).toHaveAccessibleDescription(
      '名前を入力してください',
    );
  });

  it('exposes server field errors as the input accessible description', async () => {
    const user = userEvent.setup();
    const submit = vi.fn(async () => {
      throw new ApiError(new Response(null, { status: 422 }), {
        code: 'validation_error',
        message: 'Invalid',
        errors: [{ field: 'description', message: '説明が長すぎます' }],
      });
    });
    render(<ItemForm submit={submit} />);
    await user.type(screen.getByLabelText('名前'), 'Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByLabelText('説明')).toHaveAccessibleDescription(
      '説明が長すぎます',
    );
  });

  it('shows server field errors on the matching input', async () => {
    const user = userEvent.setup();
    const submit = vi.fn(async () => {
      throw new ApiError(new Response(null, { status: 422 }), {
        code: 'validation_error',
        message: 'Invalid',
        errors: [{ field: 'name', message: '同じ名前は使えません' }],
      });
    });
    render(<ItemForm submit={submit} />);
    await user.type(screen.getByLabelText('名前'), 'Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByRole('alert')).toHaveTextContent(
      '同じ名前は使えません',
    );
  });

  it('shows unmatched server errors as form-level alerts', async () => {
    const user = userEvent.setup();
    const submit = vi.fn(async () => {
      throw new ApiError(new Response(null, { status: 422 }), {
        code: 'validation_error',
        message: 'Invalid',
        errors: [{ field: 'owner.id', message: 'unknown owner' }],
      });
    });
    render(<ItemForm submit={submit} />);
    await user.type(screen.getByLabelText('名前'), 'Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'owner.id: unknown owner',
    );
  });
});
