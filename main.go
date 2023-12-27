package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/tkeshun/golang_todo_app/config"
	"golang.org/x/sync/errgroup"
)

// run関数にhttpサーバーの処理を分離,  curl http://localhost:port番号/メッセージ
// 実行前に環境変数を設定する. PORT=18080;TODO_ENV=dev;echo $PORT;echo $TODO_ENV
// 起動方法　PORT=18080 TODO_ENV=dev go run .　もしくは.envを設定
func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("failed to terminated server: %v", err)
		os.Exit(1)
	}
}
// Configを読むように修正
// 異常時は停止せず，error型を返す．
// context.Contextの値を引数にとり，外部からのキャンセル伝播で停止する
// context.Background() で空のContextを生成
func run(ctx context.Context) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen port %d: %v", cfg.Port, err)
	}
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
