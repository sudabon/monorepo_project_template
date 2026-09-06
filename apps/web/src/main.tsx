import { bootstrap } from './bootstrap.tsx';
import './index.css';

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('#root is missing');
}
void bootstrap(rootElement);
