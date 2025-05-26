package main

import (
	"fmt"
	"os"

	"vtt2mp3/application"
	"vtt2mp3/infrastructure/google"
	"vtt2mp3/presentation"
)

// Config はアプリケーションの設定を保持する構造体
type Config struct {
	TTSService     *google.TextToSpeechService
	VTT2MP3Service *application.VTT2MP3Service
	CLI            *presentation.CLI
}

// vtt2mp3 はGoogle Cloud Text-to-Speech APIを使用してVTTファイルをMP3またはMP4ファイルに変換するコマンドラインツールです。
// 使用方法:
//
//	vtt2mp3 -i input.vtt -o output.mp3 -l ja
//	vtt2mp3 -i input.vtt -o output.mp4 -l ja
//
// フラグ:
//
//	-i string   入力VTTファイル (デフォルト "input.vtt")
//	-o string   出力ファイル (MP3またはMP4) (デフォルト "out.mp3")
//	-l string   言語コード (デフォルト "ja")
//
// 出力ファイルの拡張子が.mp4の場合、黒い背景と字幕を含む動画が生成されます。
func main() {
	config, err := initializeApp()
	if err != nil {
		handleFatalError(err)
	}

	if err := config.CLI.Run(os.Args[1:]); err != nil {
		handleFatalError(err)
	}
}

// initializeApp はアプリケーションの依存性を初期化する
func initializeApp() (*Config, error) {
	ttsService, err := google.NewTextToSpeechService()
	if err != nil {
		return nil, fmt.Errorf("テキスト読み上げサービスの初期化に失敗: %v", err)
	}

	vtt2mp3Service := application.NewVTT2MP3Service(ttsService)
	cli := presentation.NewCLI(vtt2mp3Service)

	return &Config{
		TTSService:     ttsService,
		VTT2MP3Service: vtt2mp3Service,
		CLI:            cli,
	}, nil
}

// handleFatalError はエラーを標準エラー出力に表示してプログラムを終了する
func handleFatalError(err error) {
	fmt.Fprintf(os.Stderr, "エラー: %v\n", err)

	// Google Cloud認証エラーの場合、追加のヘルプメッセージを表示
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" &&
		(err.Error() == "テキスト読み上げサービスの初期化に失敗" ||
			err.Error() == "Google Cloud認証エラー") {
		fmt.Fprintln(os.Stderr, "\nGoogle Cloud認証が設定されていない可能性があります。")
		fmt.Fprintln(os.Stderr, "README.mdの「Google Cloud認証の設定」セクションを参照して、認証を正しく設定してください。")
	}

	os.Exit(1)
}
