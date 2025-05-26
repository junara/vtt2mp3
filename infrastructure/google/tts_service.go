package google

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"vtt2mp3/domain/audio"
	"vtt2mp3/domain/tts"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// TextToSpeechService はGoogle Cloud Text-to-Speech APIを使用してtts.TextToSpeechServiceインターフェースを実装します
type TextToSpeechService struct {
	client         *texttospeech.Client
	ctx            context.Context
	audioProcessor *audio.AudioProcessor
}

// NewTextToSpeechService は新しいGoogle Cloud Text-to-Speechサービスを作成します
func NewTextToSpeechService() (*TextToSpeechService, error) {
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		// 認証エラーに関するより詳細なメッセージを提供
		if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
			return nil, fmt.Errorf("google Cloud認証エラー: GOOGLE_APPLICATION_CREDENTIALS環境変数が設定されていません。README.mdの「Google Cloud認証の設定」セクションを参照してください: %v", err)
		}
		return nil, fmt.Errorf("google Cloud Text-to-Speechクライアントの作成に失敗しました: %v", err)
	}
	return &TextToSpeechService{
		client:         client,
		ctx:            ctx,
		audioProcessor: audio.NewAudioProcessor(),
	}, nil
}

// SynthesizeSpeech はGoogle Cloud Text-to-Speech APIを使用してテキストを音声に変換します
func (s *TextToSpeechService) SynthesizeSpeech(request tts.TextToSpeechRequest) ([]byte, error) {
	// ドメインモデルをGoogle Cloud APIリクエストにマッピング
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: request.Input.Text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: request.Voice.LanguageCode,
			SsmlGender:   mapGender(request.Voice.Gender),
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: mapAudioFormat(request.AudioConfig.AudioFormat),
		},
	}

	resp, err := s.client.SynthesizeSpeech(s.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("音声合成に失敗しました: %v", err)
	}
	return resp.AudioContent, nil
}

// SynthesizeMultiple は複数のテキストをタイミング情報付きで音声に変換します
func (s *TextToSpeechService) SynthesizeMultiple(requests []tts.TextToSpeechRequest, output io.Writer) error {
	tempDir, err := s.createTempDir()
	if err != nil {
		return err
	}
	defer s.cleanupTempDir(tempDir)

	audioFiles, err := s.processRequests(requests, tempDir)
	if err != nil {
		return err
	}

	return s.mixAudioFilesWithTiming(audioFiles, requests, output)
}

// createTempDir は一時ディレクトリを作成します
func (s *TextToSpeechService) createTempDir() (string, error) {
	return s.audioProcessor.CreateTempDir()
}

// cleanupTempDir は一時ディレクトリを削除します
func (s *TextToSpeechService) cleanupTempDir(tempDir string) {
	s.audioProcessor.CleanupTempDir(tempDir)
}

// processRequests は各テキスト音声変換リクエストを処理します
func (s *TextToSpeechService) processRequests(requests []tts.TextToSpeechRequest, tempDir string) ([]string, error) {
	audioFiles := make([]string, len(requests))

	for i, req := range requests {
		// 音声を合成
		audioContent, err := s.SynthesizeSpeech(req)
		if err != nil {
			return nil, err
		}

		// 音声コンテンツを一時ファイルに保存
		audioFile := filepath.Join(tempDir, fmt.Sprintf("audio_%d.mp3", i))
		if err := os.WriteFile(audioFile, audioContent, 0644); err != nil {
			return nil, fmt.Errorf("音声ファイルの書き込みに失敗しました: %v", err)
		}

		audioFiles[i] = audioFile
	}

	return audioFiles, nil
}

// mixAudioFilesWithTiming はffmpegを使用して全ての音声ファイルを正確なタイミングで結合します
func (s *TextToSpeechService) mixAudioFilesWithTiming(audioFiles []string, requests []tts.TextToSpeechRequest, output io.Writer) error {
	if len(audioFiles) == 0 {
		return fmt.Errorf("結合する音声ファイルがありません")
	}

	// リクエストから開始時間を抽出
	startTimes := make([]time.Duration, len(requests))
	for i, req := range requests {
		startTimes[i] = req.StartTime
	}

	// オーディオプロセッサを使用して音声ファイルを結合
	return s.audioProcessor.MixAudioFilesWithTiming(audioFiles, startTimes, output)
}

// mapGender はドメインの性別をGoogle Cloud APIの性別にマッピングします
func mapGender(gender tts.VoiceGender) texttospeechpb.SsmlVoiceGender {
	switch gender {
	case tts.Male:
		return texttospeechpb.SsmlVoiceGender_MALE
	case tts.Female:
		return texttospeechpb.SsmlVoiceGender_FEMALE
	case tts.Neutral:
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	default:
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	}
}

// mapAudioFormat はドメインの音声フォーマットをGoogle Cloud APIの音声フォーマットにマッピングします
func mapAudioFormat(format tts.AudioFormat) texttospeechpb.AudioEncoding {
	switch format {
	case tts.MP3:
		return texttospeechpb.AudioEncoding_MP3
	case tts.WAV:
		return texttospeechpb.AudioEncoding_LINEAR16
	default:
		return texttospeechpb.AudioEncoding_MP3
	}
}
