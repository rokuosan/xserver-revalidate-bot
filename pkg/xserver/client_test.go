package xserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/h2non/gock"
)

func Test_newCookie(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Valid cookie",
			args:     []string{"sessionid", "abc123"},
			expected: "sessionid=abc123; Path=/; Domain=secure.xserver.ne.jp; Secure",
		},
		{
			name:     "Empty cookie",
			args:     []string{"", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newCookie(tt.args[0], tt.args[1])
			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func Test_findUniqueIdInResponse(t *testing.T) {
	tests := []struct {
		name     string
		htmlBody string
		expected UniqueID
		wantErr  bool
	}{
		{
			name: "Valid HTML with uniqid",
			htmlBody: `<html>
				<body>
					<form>
						<input type="hidden" name="uniqid" value="abc123def456" />
						<input type="submit" value="Submit" />
					</form>
				</body>
			</html>`,
			expected: UniqueID("abc123def456"),
			wantErr:  false,
		},
		{
			name: "Multiple uniqid inputs - should return last one",
			htmlBody: `<html>
				<body>
					<form>
						<input type="hidden" name="uniqid" value="first123" />
						<input type="hidden" name="uniqid" value="second456" />
					</form>
				</body>
			</html>`,
			expected: UniqueID("second456"),
			wantErr:  false,
		},
		{
			name: "No uniqid input found",
			htmlBody: `<html>
				<body>
					<form>
						<input type="hidden" name="token" value="abc123" />
						<input type="submit" value="Submit" />
					</form>
				</body>
			</html>`,
			expected: UniqueID(""),
			wantErr:  true,
		},
		{
			name: "Empty uniqid value",
			htmlBody: `<html>
				<body>
					<form>
						<input type="hidden" name="uniqid" value="" />
					</form>
				</body>
			</html>`,
			expected: UniqueID(""),
			wantErr:  true,
		},
		{
			name: "uniqid input without value attribute",
			htmlBody: `<html>
				<body>
					<form>
						<input type="hidden" name="uniqid" />
					</form>
				</body>
			</html>`,
			expected: UniqueID(""),
			wantErr:  true,
		},
		{
			name:     "Invalid HTML",
			htmlBody: `<invalid><html>`,
			expected: UniqueID(""),
			wantErr:  true,
		},
		{
			name: "Complex HTML with nested elements",
			htmlBody: `<!DOCTYPE html>
			<html>
				<head><title>Test</title></head>
				<body>
					<div class="container">
						<form method="post" action="/submit">
							<div class="form-group">
								<input type="text" name="username" />
							</div>
							<input type="hidden" name="csrf_token" value="other123" />
							<input type="hidden" name="uniqid" value="complex789xyz" />
							<div class="form-group">
								<input type="password" name="password" />
							</div>
						</form>
					</div>
				</body>
			</html>`,
			expected: UniqueID("complex789xyz"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.htmlBody)
			result, err := findUniqueIdInResponse(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func Test_findErrorMessageFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		htmlBody string
		expected []string
		wantErr  bool
	}{
		{
			name: "Single error message in main .contents",
			htmlBody: `<html>
				<body>
					<main>
						<div class="contents">エラーが発生しました。</div>
					</main>
				</body>
			</html>`,
			expected: []string{"エラーが発生しました。"},
			wantErr:  false,
		},
		{
			name: "Multiple error messages in main .contents",
			htmlBody: `<html>
				<body>
					<main>
						<div class="contents">エラー1: 無効な入力です。</div>
						<div class="contents">エラー2: セッションが無効です。</div>
					</main>
				</body>
			</html>`,
			expected: []string{"エラー1:", "無効な入力です。", "エラー2:", "セッションが無効です。"},
			wantErr:  false,
		},
		{
			name: "Error message with whitespace and line breaks",
			htmlBody: `<html>
				<body>
					<main>
						<div class="contents">
							エラーが発生しました。
							再試行してください。
						</div>
					</main>
				</body>
			</html>`,
			expected: []string{"エラーが発生しました。", "再試行してください。"},
			wantErr:  false,
		},
		{
			name: "No main .contents elements",
			htmlBody: `<html>
				<body>
					<div class="contents">これはmainの中にありません</div>
					<main>
						<div class="other">これは.contentsではありません</div>
					</main>
				</body>
			</html>`,
			expected: []string{"これは.contentsではありません"},
			wantErr:  false,
		},
		{
			name: "Empty .contents elements",
			htmlBody: `<html>
				<body>
					<main>
						<div class="contents"></div>
						<div class="contents">   </div>
					</main>
				</body>
			</html>`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Mixed empty and non-empty .contents",
			htmlBody: `<html>
				<body>
					<main>
						<div class="contents"></div>
						<div class="contents">有効なエラーメッセージ</div>
						<div class="contents">   </div>
					</main>
				</body>
			</html>`,
			expected: []string{"有効なエラーメッセージ"},
			wantErr:  false,
		},
		{
			name: "Nested elements within .contents",
			htmlBody: `<html>
				<body>
					<main>
						<div class="contents">
							<p>段落1: エラーです</p>
							<span>スパン: 詳細情報</span>
						</div>
					</main>
				</body>
			</html>`,
			expected: []string{"段落1:", "エラーです", "スパン:", "詳細情報"},
			wantErr:  false,
		},
		{
			name:     "Invalid HTML",
			htmlBody: `<invalid><html>`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Complex HTML structure with multiple main elements",
			htmlBody: `<!DOCTYPE html>
			<html>
				<head><title>Error Page</title></head>
				<body>
					<header>ヘッダー</header>
					<main>
						<div class="contents">メインエラー: システムエラーが発生しました</div>
					</main>
					<main>
						<div class="contents">セカンダリエラー: データベース接続エラー</div>
					</main>
					<footer>フッター</footer>
				</body>
			</html>`,
			expected: []string{"メインエラー:", "システムエラーが発生しました", "セカンダリエラー:", "データベース接続エラー"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.htmlBody)
			result, err := findErrorMessageFromResponse(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func Test_GetCSRFTokenAsUniqueID(t *testing.T) {
	defer gock.Off()

	t.Run("Should return UniqueID from response", func(t *testing.T) {
		gock.New("https://secure.xserver.ne.jp").
			Get("/xapanel/xvps/server/freevps/extend/index").
			Reply(200).
			BodyString(`<html>
			<body>
				<form>
					<input type="hidden" name="uniqid" value="csrf1234567890" />
					<input type="submit" value="Submit" />
				</form>
			</body>
		</html>`)
		c := &client{
			Client: &http.Client{},
			Headers: map[string]string{
				"User-Agent": "TestClient/1.0",
			},
			Logger: slog.Default(),
		}
		ctx := context.Background()

		vpsID := VPSID("test-vps-id")
		uniqueID, err := c.GetCSRFTokenAsUniqueID(ctx, vpsID)
		if err != nil {
			t.Fatalf("GetCSRFTokenAsUniqueID failed: %v", err)
		}
		expected := UniqueID("csrf1234567890")
		if uniqueID != expected {
			t.Errorf("expected %s, got %s", expected, uniqueID)
		}
	})
}

func Test_ExtendFreeVPSExpiration(t *testing.T) {
	defer gock.Off()

	t.Run("Should extend free VPS expiration successfully", func(t *testing.T) {
		vpsID := VPSID("test-vps-id")
		uniqueID := UniqueID("csrf1234567890")
		formData := fmt.Sprintf("uniqid=%s&ethna_csrf=&id_vps=%s", uniqueID, vpsID)

		gock.New("https://" + XServerHost).
			Post(DoFreeVPSExtendPath).
			Body(strings.NewReader(formData)).
			Reply(200).
			BodyString(`<html>
			<body>
				<p>利用期限の更新手続きが完了しました。</p>
			</body>
		</html>`)

		c := &client{
			Client: &http.Client{},
			Headers: map[string]string{
				"User-Agent": "TestClient/1.0",
			},
			Logger: slog.Default(),
		}
		ctx := context.Background()

		err := c.ExtendFreeVPSExpiration(ctx, vpsID, uniqueID)
		if err != nil {
			t.Fatalf("ExtendFreeVPSExpiration failed: %v", err)
		}
	})

	t.Run("Should return error on failed extension", func(t *testing.T) {
		vpsID := VPSID("test-vps-id")
		uniqueID := UniqueID("csrf1234567890")
		formData := fmt.Sprintf("uniqid=%s&ethna_csrf=&id_vps=%s", uniqueID, vpsID)

		gock.New("https://" + XServerHost).
			Post(DoFreeVPSExtendPath).
			Body(strings.NewReader(formData)).
			Reply(200).
			BodyString(`<html>
			<body>
				<main>
					<div class="contents">This is an error</div>
				</main>
			</body>
		</html>`)

		c := &client{
			Client: &http.Client{},
			Headers: map[string]string{
				"User-Agent": "TestClient/1.0",
			},
			Logger: slog.Default(),
		}
		ctx := context.Background()

		err := c.ExtendFreeVPSExpiration(ctx, vpsID, uniqueID)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		expectedError := "VPS renewal failed: This is an error"
		if err.Error() != expectedError {
			t.Errorf("expected %s, got %s", expectedError, err.Error())
		}
	})
}
