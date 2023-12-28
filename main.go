package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

// run関数にhttpサーバーの処理を分離,  curl http://localhost:port番号/メッセージ
func main() {
	// 引数でポート番号を指定できるように変更
	fmt.Println(os.Args)
	// 引数が読み込まれないのでとりあえず無効化
	// if len(os.Args) != 2 {
	// 	log.Printf("need port number\n")
	// 	os.Exit(1)
	// }
	// p := os.Args[1]
	p := "80"
	l, err := net.Listen("tcp", ":"+p)
	if err != nil {
		log.Fatalf("failed to listen port %s: %v", p, err)
	}
	// すべての goroutine が終わるのを待つ
	// エラーハンドリング
	if err := run(context.Background(), l); err != nil {
		log.Printf("failed to terminate server: %v", err)
		os.Exit(1)
	}
}

// 異常時は停止せず，error型を返す．
// context.Contextの値を引数にとり，外部からのキャンセル伝播で停止する
// context.Background() で空のContextを生成
func run(ctx context.Context, l net.Listener) error {
	// http.Server型．ListenAndServeメソッドでサーバーを起動する．
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
		}),
	}

	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでHTTPサーバーを起動する
	eg.Go(func() error {
		// httpリクエストを受け付ける
		// ErrServerClosedはサーバーが正常終了したことを示すのでエラーハンドリングから除外
		// Serveの引数にListenerを指定
		if err := s.Serve(l); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("failed to close: %+v", err)
			// eg.Go() で実行された関数にエラーを返すものがいれば、ctx はキャンセルされる
			return err
		}
		return nil
	})
	// Done通知があったらサーバーを終了させる
	<-ctx.Done()
	// context.Background() で空のContextを生成
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}
	// Goメソッドで起動した別ゴルーチンの終了を待つ
	return eg.Wait()
}
