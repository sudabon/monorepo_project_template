import { runtimeConfigSchema, type RuntimeConfig } from './schema.ts';

export class ConfigLoadError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ConfigLoadError';
  }
}

export async function loadRuntimeConfig(
  fetchImpl: typeof fetch = fetch,
): Promise<RuntimeConfig> {
  let response: Response;
  try {
    response = await fetchImpl('/config.json', { cache: 'no-store' });
  } catch {
    throw new ConfigLoadError('設定ファイルを読み込めませんでした');
  }
  if (!response.ok) {
    throw new ConfigLoadError('設定ファイルを読み込めませんでした');
  }
  let json: unknown;
  try {
    json = await response.json();
  } catch {
    throw new ConfigLoadError('設定ファイルの形式が不正です');
  }
  const parsed = runtimeConfigSchema.safeParse(json);
  if (!parsed.success) {
    throw new ConfigLoadError('設定ファイルの形式が不正です');
  }
  return parsed.data;
}
