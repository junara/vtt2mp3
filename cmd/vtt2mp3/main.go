package main

import (
	"fmt"
	"os"
	"vtt2mp3/application"
	"vtt2mp3/infrastructure/google"
	"vtt2mp3/presentation"
)

// exitCode はプログラムの終了コードを表す
const exitCode = 1

// handleError はエラーを標準エラー出力に表示し、指定された終了コードでプログラムを終了する
func handleError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)

	// Google Cloud認証エラーの場合、追加のヘルプメッセージを表示
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" &&
		(err.Error() == "failed to initialize TTS service" ||
			err.Error() == "Google Cloud認証エラー") {
		fmt.Fprintln(os.Stderr, "\nGoogle Cloud認証が設定されていない可能性があります。")
		fmt.Fprintln(os.Stderr, "README.mdの「Google Cloud認証の設定」セクションを参照して、認証を正しく設定してください。")
	}

	os.Exit(exitCode)
}

// initializeApp はアプリケーションの依存関係を初期化し、CLIインターフェースを返す
func initializeApp() (*presentation.CLI, error) {
	// Google Cloud Text-to-Speechサービスの作成
	ttsService, err := google.NewTextToSpeechService()
	if err != nil {
		return nil, fmt.Errorf("TTSサービスの初期化に失敗しました: %w", err)
	}

	// アプリケーションサービスの作成
	vtt2mp3Service := application.NewVTT2MP3Service(ttsService)

	// CLIの作成
	cli := presentation.NewCLI(vtt2mp3Service)

	return cli, nil
}

func main() {
	// アプリケーションの初期化
	cli, err := initializeApp()
	if err != nil {
		handleError(err)
	}

	// CLIの実行
	if err := cli.Run(os.Args[1:]); err != nil {
		handleError(err)
	}
}
