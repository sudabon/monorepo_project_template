import { createFileRoute } from '@tanstack/react-router';
import { HomePage } from '../../pages/HomePage.tsx';

export const Route = createFileRoute('/_authenticated/')({
  component: HomePage,
});
