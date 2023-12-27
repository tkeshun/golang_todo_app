package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestRun(t *testing.T) {
	// キャンセル可能なcontextオブジェクトを作る
	// context自体は値が空のコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	// errgroupで戻り値にエラーを含むゴルーチンを扱いやすくする
	eg, ctx := errgroup.WithContext(ctx)
	// ゴルーチンでhttpサーバーの起動
	eg.Go(func() error {
		return run(ctx)
	})
	// httpリクエストの送信 
	in := "message"
	rsp, err := http.Get("http://localhost:18080/"+in)
	// 送信のエラーハンドリング
	if err != nil {
		t.Errorf("failed to get: %+v", err)
	}
	// HTTPレスポンスのボディ (rsp.Body) をクローズする
	// 関数が終了する前に確実にクローズ処理が行われるようにする
	defer rsp.Body.Close()
	// Bodyから値を読み出す
	got ,err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	// 期待値と実際の値の比較
	want := fmt.Sprintf("Hello, %s!",in)
	if string(got) != want {
		t.Errorf("want %q,but got %q",want,got)
	}
	// キャンセルとゴルーチンの終了待ち
	cancel()
	// run関数の返り値を検証する
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}