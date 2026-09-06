export function ConfigErrorScreen({ message }: { message: string }) {
  return (
    <main className="mx-auto max-w-lg p-8">
      <h1 className="text-xl font-semibold">設定を読み込めませんでした</h1>
      <p className="mt-3 text-muted-foreground">{message}</p>
    </main>
  );
}
