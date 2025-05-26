package tts

import (
	"io"
	"time"
)

// AudioFormat は音声出力のフォーマットを表します
type AudioFormat int

const (
	// MP3 はMP3音声フォーマットを表します
	MP3 AudioFormat = iota
	// WAV はWAV音声フォーマットを表します
	WAV
)

// String はAudioFormatを文字列に変換します
func (f AudioFormat) String() string {
	switch f {
	case MP3:
		return "MP3"
	case WAV:
		return "WAV"
	default:
		return "UNKNOWN"
	}
}

// VoiceGender は音声の性別を表します
type VoiceGender int

const (
	// Male は男性の声を表します
	Male VoiceGender = iota
	// Female は女性の声を表します
	Female
	// Neutral は中性的な声を表します
	Neutral
)

// String はVoiceGenderを文字列に変換します
func (g VoiceGender) String() string {
	switch g {
	case Male:
		return "MALE"
	case Female:
		return "FEMALE"
	case Neutral:
		return "NEUTRAL"
	default:
		return "UNKNOWN"
	}
}

// SynthesisInput は音声合成の入力テキストを表します
type SynthesisInput struct {
	// Text は音声に変換するテキスト内容
	Text string
}

// VoiceSelectionParams は音声選択パラメータを表します
type VoiceSelectionParams struct {
	// LanguageCode は言語コード（例: "ja-JP", "en-US"）
	LanguageCode string
	// Gender は声の性別
	Gender VoiceGender
}

// AudioConfig は音声出力の設定を表します
type AudioConfig struct {
	// AudioFormat は音声のフォーマット（MP3, WAVなど）
	AudioFormat AudioFormat
}

// TextToSpeechRequest はテキストから音声への変換リクエストを表します
type TextToSpeechRequest struct {
	// Input は合成する入力テキスト
	Input SynthesisInput
	// Voice は使用する音声の選択
	Voice VoiceSelectionParams
	// AudioConfig は出力音声の設定
	AudioConfig AudioConfig
	// StartTime は複数テキストを合成する際の開始時間
	StartTime time.Duration
}

// TextToSpeechService はテキスト読み上げサービスのインターフェースを定義します
type TextToSpeechService interface {
	// SynthesizeSpeech はテキストを音声に変換し、音声コンテンツを返します
	SynthesizeSpeech(request TextToSpeechRequest) ([]byte, error)

	// SynthesizeMultiple は複数のテキストをタイミング情報付きで音声に変換します
	SynthesizeMultiple(requests []TextToSpeechRequest, output io.Writer) error
}
