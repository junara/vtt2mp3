package application

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"vtt2mp3/domain/tts"
	"vtt2mp3/domain/vtt"
)

// エラーメッセージの定数
const (
	errParseVTT     = "VTTファイルの解析に失敗: %w"
	errCreateOutput = "出力ファイルの作成に失敗: %w"
	errCloseOutput  = "出力ファイルの閉じるのに失敗: %w"
	errSynthesize   = "音声合成に失敗: %w"
	errCreateVideo  = "動画作成に失敗: %w"
)

// VTT2MP3Service は字幕ファイル(VTT)からMP3音声ファイルへの変換を行うサービス
type VTT2MP3Service struct {
	ttsService tts.TextToSpeechService
}

// NewVTT2MP3Service はVTT2MP3Serviceの新しいインスタンスを作成する
func NewVTT2MP3Service(ttsService tts.TextToSpeechService) *VTT2MP3Service {
	return &VTT2MP3Service{
		ttsService: ttsService,
	}
}

// ConvertOptions はVTTからMP3またはMP4への変換オプションを表す
type ConvertOptions struct {
	InputFile     string // 入力VTTファイルのパス
	OutputFile    string // 出力MP3またはMP4ファイルのパス
	LanguageCode  string // 音声合成に使用する言語コード
	IsVideoOutput bool   // 出力が動画かどうか
}

// Convert はVTTファイルをMP3ファイルまたはMP4ファイルに変換する
func (s *VTT2MP3Service) Convert(options ConvertOptions) error {
	// VTTファイルを解析
	vttFile, err := vtt.ParseVTTFile(options.InputFile)
	if err != nil {
		return fmt.Errorf(errParseVTT, err)
	}

	// 動画出力の場合
	if options.IsVideoOutput {
		return s.convertToVideo(vttFile, options)
	}

	// 音声出力の場合（MP3）
	return s.convertToAudio(vttFile, options)
}

// convertToAudio はVTTファイルをMP3ファイルに変換する
func (s *VTT2MP3Service) convertToAudio(vttFile *vtt.VTTFile, options ConvertOptions) error {
	// 字幕からTTSリクエストを作成
	ttsRequests := s.createTTSRequests(vttFile, options.LanguageCode)

	// 出力ファイルを作成
	outputFile, err := os.Create(options.OutputFile)
	if err != nil {
		return fmt.Errorf(errCreateOutput, err)
	}
	defer func() {
		closeErr := outputFile.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf(errCloseOutput, closeErr)
		}
	}()

	// 音声を合成して出力ファイルに書き込む
	if err := s.ttsService.SynthesizeMultiple(ttsRequests, outputFile); err != nil {
		return fmt.Errorf(errSynthesize, err)
	}

	return nil
}

// convertToVideo はVTTファイルをMP4動画ファイルに変換する
func (s *VTT2MP3Service) convertToVideo(vttFile *vtt.VTTFile, options ConvertOptions) error {
	// 一時的なMP3ファイルを作成
	tempDir, err := os.MkdirTemp("", "vtt2mp4_")
	if err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("一時ディレクトリの削除に失敗しました: %v\n", err)
		}
	}()

	// 一時的なMP3ファイルのパス
	tempMP3 := filepath.Join(tempDir, "audio.mp3")

	// 一時的なVTTファイルのパス
	tempVTT := filepath.Join(tempDir, "subtitles.vtt")

	// 一時的なMP3ファイルを作成するためのオプション
	audioOptions := ConvertOptions{
		InputFile:     options.InputFile,
		OutputFile:    tempMP3,
		LanguageCode:  options.LanguageCode,
		IsVideoOutput: false,
	}

	// 音声を生成
	if err := s.convertToAudio(vttFile, audioOptions); err != nil {
		return fmt.Errorf("音声生成に失敗: %w", err)
	}

	// 一時的なVTTファイルを作成
	if err := s.writeVTTFile(vttFile, tempVTT); err != nil {
		return fmt.Errorf("字幕ファイルの作成に失敗: %w", err)
	}

	// FFmpegを使用して動画を生成
	if err := s.generateVideo(tempMP3, tempVTT, options.OutputFile); err != nil {
		return fmt.Errorf(errCreateVideo, err)
	}

	return nil
}

// createTTSRequests は字幕データからTTSリクエストのスライスを作成する
func (s *VTT2MP3Service) createTTSRequests(vttFile *vtt.VTTFile, languageCode string) []tts.TextToSpeechRequest {
	ttsRequests := make([]tts.TextToSpeechRequest, 0, len(vttFile.Subtitles))

	for _, subtitle := range vttFile.Subtitles {
		ttsRequests = append(ttsRequests, tts.TextToSpeechRequest{
			Input: tts.SynthesisInput{
				Text: subtitle.Text,
			},
			Voice: tts.VoiceSelectionParams{
				LanguageCode: languageCode,
				Gender:       tts.Neutral,
			},
			AudioConfig: tts.AudioConfig{
				AudioFormat: tts.MP3,
			},
			StartTime: subtitle.StartTime,
		})
	}

	return ttsRequests
}

// writeVTTFile はVTTファイルを指定されたパスに書き込む
func (s *VTT2MP3Service) writeVTTFile(vttFile *vtt.VTTFile, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("VTTファイルの作成に失敗: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("ファイルの閉じるのに失敗しました: %v\n", err)
		}
	}()

	// VTTヘッダーを書き込む
	if _, err := file.WriteString("WEBVTT\n\n"); err != nil {
		return fmt.Errorf("VTTヘッダーの書き込みに失敗: %w", err)
	}

	// 各字幕を書き込む
	for i, subtitle := range vttFile.Subtitles {
		// タイムスタンプを書き込む
		startTime := formatTimestamp(subtitle.StartTime)
		endTime := formatTimestamp(subtitle.EndTime)

		if _, err := fmt.Fprintf(file, "%d\n%s --> %s\n%s\n\n",
			i+1, startTime, endTime, subtitle.Text); err != nil {
			return fmt.Errorf("字幕の書き込みに失敗: %w", err)
		}
	}

	return nil
}

// formatTimestamp は時間をVTTのタイムスタンプ形式（HH:MM:SS.mmm）に変換する
func formatTimestamp(duration time.Duration) string {
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	milliseconds := int(duration.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

// generateVideo はMP3音声ファイルとVTT字幕ファイルからMP4動画を生成する
func (s *VTT2MP3Service) generateVideo(audioFile, subtitleFile, outputFile string) error {
	// FFmpegコマンドを構築
	// 1. 黒い背景の動画を生成
	// 2. 音声ファイルを追加
	// 3. 字幕を追加（画面上部に表示）
	// 4. タイムコードを追加（画面中央に表示、0.1秒単位）

	// 複数のフィルターを組み合わせる
	// - subtitles: 字幕を追加（上部中央に表示）
	// - drawtext: タイムコードを表示（中央に表示、0.1秒単位）
	// タイムコード表示のカスタムフォーマット（HH:MM:SS.T形式、0.1秒単位）
	// FFmpegのtextフィルターで時間を表示する際に、0.1秒単位で表示するようにカスタマイズ
	videoFilter := "subtitles=" + subtitleFile + ":force_style='Alignment=6,FontSize=24'," +
		"drawtext=fontsize=48:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2:" +
		"text='%{pts\\:hms}.%{eif\\:mod(floor(t*10),10)\\:d}':box=1:boxcolor=black@0.5:boxborderw=5:rate=10"

	cmd := exec.Command(
		"ffmpeg",
		"-y",          // 既存のファイルを上書き
		"-f", "lavfi", // 入力フォーマットとしてlavfiを使用
		"-i", "color=c=black:s=1280x720:r=30", // 黒い背景の1280x720、30fpsの動画を生成
		"-i", audioFile, // 音声ファイルを入力として追加
		"-vf", videoFilter, // 字幕とタイムコードを追加
		"-c:a", "aac", // 音声コーデックとしてAACを使用
		"-c:v", "libx264", // 動画コーデックとしてH.264を使用
		"-shortest", // 最も短い入力の長さに合わせる
		outputFile,  // 出力ファイル
	)

	// コマンドを実行
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpegの実行に失敗: %w, 出力: %s", err, string(output))
	}

	return nil
}
