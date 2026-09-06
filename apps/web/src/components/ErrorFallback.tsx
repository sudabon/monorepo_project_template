export function RouteErrorFallback({
  error,
}: {
  error?: { message?: string };
}) {
  return (
    <section className="p-6" role="alert">
      <h2 className="text-lg font-semibold">
        この画面の表示中にエラーが発生しました
      </h2>
      <p className="mt-2 text-muted-foreground">
        {error?.message ?? '再読み込みするか、別の画面へ移動してください。'}
      </p>
    </section>
  );
}

export function RootErrorFallback() {
  return (
    <main className="p-8" role="alert">
      <h1 className="text-xl font-semibold">
        アプリケーションでエラーが発生しました
      </h1>
      <p className="mt-2 text-muted-foreground">
        ページを再読み込みしてください。
      </p>
    </main>
  );
}
