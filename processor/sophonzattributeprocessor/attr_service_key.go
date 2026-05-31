// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"

	sophonzmetadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

type ServiceKey struct {
	ServiceID   string `json:"sub"`
	ServiceName string `json:"serviceName"`
	ExpiredAt   time.Time
}

func (p *SOPHONZAttributeProcessor) validateServiceKeyAndSetServiceResource(attrs pcommon.Map) bool {
	value, exist := attrs.Get(sophonzsemconv.ServiceKey)
	if !exist {
		return true
	}
	result, err := decrypt(value.Str())
	if err != nil {
		return true
	}

	var serviceKey ServiceKey
	if err = json.Unmarshal([]byte(result), &serviceKey); err != nil {
		p.logger.Error("failed to unmarshal service key", zap.Error(err))
		return true
	}
	//todo 서비스키 못찾을 시 수집할지 거부할지
	if p.MetadataManager == nil {
		p.logger.Warn("metadataManager is not initialized")
		return true
	}

	skm, ok := p.MetadataManager.Service.Load().(sophonzmetadata.ServiceKeyMap)
	if !ok || skm == nil {
		p.logger.Error("service key map is not initialized")
		return true
	}
	service, ok := skm[serviceKey.ServiceID]
	if !ok {
		return true
	}
	if serviceKey.ServiceID != service.ID || service.Name != serviceKey.ServiceName {
		return true
	}

	return false
}

// ENCRYPTION_KEY는 환경 변수 또는 기본값으로 설정됩니다.
var ENCRYPTION_KEY = []byte("api-secret-token-random-string--")

func encrypt(payload string) (string, error) {
	// IV (초기화 벡터) 생성
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	block, err := aes.NewCipher(ENCRYPTION_KEY)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// CTR 암호화 스트림 생성
	stream := cipher.NewCTR(block, iv)
	plaintext := []byte(payload)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// 결과를 base64url로 인코딩
	encrypted := base64.RawURLEncoding.EncodeToString(ciphertext)
	ivEncoded := base64.RawURLEncoding.EncodeToString(iv)
	return fmt.Sprintf("%s.%s", encrypted, ivEncoded), nil
}

// Decrypt는 암호화된 payload를 복호화합니다.
func decrypt(encryptedPayload string) (string, error) {
	// 암호화된 텍스트와 IV를 "." 기준으로 분리
	parts := bytes.Split([]byte(encryptedPayload), []byte("."))
	if len(parts) != 2 {
		return "", errors.New("invalid payload format")
	}

	encryptedText, ivText := parts[0], parts[1]

	// Base64url 디코딩
	ciphertext, err := base64.RawURLEncoding.DecodeString(string(encryptedText))
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %v", err)
	}

	iv, err := base64.RawURLEncoding.DecodeString(string(ivText))
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %v", err)
	}

	block, err := aes.NewCipher(ENCRYPTION_KEY)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// CTR 복호화 스트림 생성
	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}
