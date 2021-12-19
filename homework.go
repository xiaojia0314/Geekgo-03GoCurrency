package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"os/signal"
)

func Server(w http.ResponseWriter, r *http.Request){
	io.WriteString(w,"hello Server")
}


func StartHttpServer(server *http.Server)error{
	http.HandleFunc("/server", Server)
	err := server.ListenAndServe()
	return err
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	group, errCtx := errgroup.WithContext(ctx)

	srv := &http.Server{Addr: "8080"}

	group.Go(func() error {
		return StartHttpServer(srv)
	})


	group.Go(func() error {
		<-errCtx.Done() //阻塞。因为 cancel、timeout、deadline 都可能导致 Done 被 close
		fmt.Println("http server stop")
		return srv.Shutdown(errCtx) // 关闭 http server
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	group.Go(func() error {
		for{
			select {
			case <- errCtx.Done():
				return  errCtx.Err()
			case <- sig:
				cancel()
			}
		}
		return nil
	})

	if err := group.Wait(); err!=nil{
		fmt.Println("Group error ", err)
	}
	fmt.Println("All done")

}