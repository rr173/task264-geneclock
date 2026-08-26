// 合成生物基因调控时序复核台入口。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task264-geneclock/internal/httpapi"
	"task264-geneclock/internal/service"
	"task264-geneclock/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath := flag.String("db", "geneclock.db", "SQLite 数据库路径")
	smoke := flag.Bool("smoke-test", false, "运行端到端自检后退出")
	flag.Parse()

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer st.Close()

	app := service.New(st)

	if *smoke {
		if err := smokeTest(app); err != nil {
			fmt.Fprintf(os.Stderr, "smoke-test FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("smoke-test PASSED")
		return
	}

	srv := &http.Server{
		Addr:    *addr,
		Handler: httpapi.New(app).Handler(),
	}

	go func() {
		log.Printf("基因调控时序复核台监听 %s", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务错误: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("服务已退出")
}
