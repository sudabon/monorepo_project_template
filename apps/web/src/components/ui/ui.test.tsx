import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Button } from './button.tsx';
import { Input } from './input.tsx';
import { Modal } from './modal.tsx';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from './table.tsx';
import { ToastProvider, ToastViewport, useToast } from './toast.tsx';

describe('ui components', () => {
  it('renders a button that can be pressed', async () => {
    const onClick = vi.fn();
    const user = userEvent.setup();
    render(<Button onClick={onClick}>保存</Button>);
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('associates an input with its label', async () => {
    const user = userEvent.setup();
    render(<Input label="名前" />);
    await user.type(screen.getByLabelText('名前'), 'Widget');
    expect(screen.getByLabelText('名前')).toHaveValue('Widget');
  });

  it('renders table content', () => {
    render(
      <Table>
        <TableHead>
          <TableRow>
            <TableHeaderCell>名前</TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow>
            <TableCell>Widget</TableCell>
          </TableRow>
        </TableBody>
      </Table>,
    );
    expect(
      screen.getByRole('columnheader', { name: '名前' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: 'Widget' })).toBeInTheDocument();
  });

  it('traps focus and closes the modal with Escape', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    function Harness() {
      return (
        <>
          <button type="button">外側</button>
          <Modal open title="確認" onClose={onClose}>
            <button type="button">中の操作</button>
          </Modal>
        </>
      );
    }
    render(<Harness />);
    expect(screen.getByRole('dialog', { name: '確認' })).toBeInTheDocument();
    await user.keyboard('{Escape}');
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('shows a toast message', async () => {
    const user = userEvent.setup();
    function Trigger() {
      const { pushToast } = useToast();
      return (
        <button
          type="button"
          onClick={() => pushToast({ message: '保存しました' })}
        >
          通知
        </button>
      );
    }
    render(
      <ToastProvider>
        <Trigger />
        <ToastViewport />
      </ToastProvider>,
    );
    await user.click(screen.getByRole('button', { name: '通知' }));
    expect(screen.getByRole('status')).toHaveTextContent('保存しました');
  });
});
