package validator

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"github.com/grigri/pv/internal/errors"
)

// yamlValidatorImpl implements the YAMLValidator interface
const (
	// 安全限制常量
	MaxFrontMatterLines = 500      // 最大 front matter 行数
	MaxFileSize         = 10 << 20 // 10MB 最大文件大小
	MaxLineLength       = 4096     // 最大行长度
)

// FrontMatterParser 健壮的 Front Matter 解析器
type FrontMatterParser struct{}

// ParseResult Front Matter 解析结果
type ParseResult struct {
	YAMLContent string
	BodyContent string
}

// Parse 解析 front matter 格式的内容
func (p *FrontMatterParser) Parse(content []byte) (*ParseResult, error) {
	if len(content) == 0 {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"empty content",
			errors.ErrInvalidYAML,
		)
	}

	if len(content) > MaxFileSize {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			fmt.Sprintf("file too large: %d bytes (max: %d)", len(content), MaxFileSize),
			errors.ErrInvalidYAML,
		)
	}

	// 检查是否为有效的 UTF-8
	if !utf8.Valid(content) {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"invalid UTF-8 content",
			errors.ErrInvalidYAML,
		)
	}

	return p.parseContent(content)
}

// parseContent 执行实际的解析逻辑
func (p *FrontMatterParser) parseContent(content []byte) (*ParseResult, error) {
	// 将内容转换为字符串进行处理
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	if len(lines) == 0 {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"empty content",
			errors.ErrInvalidYAML,
		)
	}

	var frontMatterLines []string
	var contentLines []string
	var frontMatterClosed bool
	separatorIndex := -1

	// 检查第一行
	firstLine := strings.TrimSpace(lines[0])

	if firstLine == "---" {
		// 标准格式：以 --- 开头
		// 寻找结束的 ---
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				separatorIndex = i
				frontMatterClosed = true
				break
			}
			frontMatterLines = append(frontMatterLines, lines[i])
		}
	} else {
		// 非标准格式：寻找第一个 --- 作为分隔符
		for i := 0; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				separatorIndex = i
				break
			}
			frontMatterLines = append(frontMatterLines, lines[i])
		}

		if separatorIndex != -1 {
			frontMatterClosed = true
		}
	}

	// 处理特殊情况
	if len(lines) == 1 && firstLine == "---" {
		// 只有一个 --- 的情况
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"unclosed front matter: missing closing '---'",
			errors.ErrInvalidYAML,
		)
	}

	// 如果没有找到分隔符
	if separatorIndex == -1 && firstLine != "---" {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"YAML front matter separator '---' not found",
			errors.ErrInvalidYAML,
		)
	}

	// 如果 front matter 没有正确闭合
	if !frontMatterClosed {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"unclosed front matter: missing closing '---'",
			errors.ErrInvalidYAML,
		)
	}

	// 提取内容部分
	if separatorIndex != -1 && separatorIndex < len(lines)-1 {
		contentLines = lines[separatorIndex+1:]
	}

	// 处理 YAML 内容
	yamlContent := strings.Join(frontMatterLines, "\n")
	yamlContent = strings.TrimSpace(yamlContent)

	// 对于 ---\n--- 这种空的 front matter，允许通过但设置为空
	if yamlContent == "" && firstLine == "---" {
		// 允许空的标准格式 front matter，但返回空字符串
		yamlContent = ""
	} else if yamlContent == "" {
		return nil, errors.NewAppError(
			errors.ErrValidation,
			"empty front matter content",
			errors.ErrInvalidYAML,
		)
	}

	// 处理内容部分
	bodyContent := strings.Join(contentLines, "\n")
	bodyContent = strings.TrimSpace(bodyContent)

	return &ParseResult{
		YAMLContent: yamlContent,
		BodyContent: bodyContent,
	}, nil
}

// looksLikeYAML 简单启发式检查是否像 YAML 内容
func (p *FrontMatterParser) looksLikeYAML(line string) bool {
	if line == "" {
		return false
	}

	// 检查是否以 # 开头（YAML 注释）
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return true
	}

	// 检查常见的 YAML 模式
	patterns := []string{
		"name:", "title:", "author:", "description:",
		"tags:", "version:", "date:", "draft:",
	}

	lowered := strings.ToLower(line)
	for _, pattern := range patterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}

	// 检查是否是 key: value 格式
	if strings.Contains(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
		return true
	}

	return false
}

type yamlValidatorImpl struct{}

// NewYAMLValidator creates a new YAML validator instance
func NewYAMLValidator() YAMLValidator {
	return &yamlValidatorImpl{}
}

// ValidatePromptFile validates a YAML prompt file and returns parsed content
func (v *yamlValidatorImpl) ValidatePromptFile(content []byte) (*PromptFileContent, error) {
	// 使用健壮的 Front Matter 解析器
	parser := &FrontMatterParser{}

	result, err := parser.Parse(content)
	if err != nil {
		return nil, err // 错误已经被包装过了
	}

	// 处理空的 YAML 内容（如 ---\n--- 的情况）
	var metadata PromptMetadata
	if strings.TrimSpace(result.YAMLContent) != "" {
		if err := yaml.Unmarshal([]byte(result.YAMLContent), &metadata); err != nil {
			return nil, errors.NewAppError(
				errors.ErrValidation,
				"failed to parse YAML metadata",
				errors.ErrInvalidYAML,
			)
		}
	}
	// 如果 YAML 内容为空，metadata 将保持零值状态

	return &PromptFileContent{
		Metadata: metadata,
		Content:  result.BodyContent,
	}, nil
}

// ValidateRequired validates that required fields are present and valid
func (v *yamlValidatorImpl) ValidateRequired(prompt *PromptFileContent) error {
	if prompt == nil {
		return errors.NewAppError(
			errors.ErrValidation,
			"prompt content cannot be nil",
			errors.ErrInvalidMetadata,
		)
	}

	// Validate required field: name
	if strings.TrimSpace(prompt.Metadata.Name) == "" {
		return errors.ValidationError{
			Field:   "name",
			Message: "name is required and cannot be empty",
		}
	}

	// Validate name length (1-100 characters as per design doc)
	if len(prompt.Metadata.Name) > 100 {
		return errors.ValidationError{
			Field:   "name",
			Message: "name cannot be longer than 100 characters",
		}
	}

	// Validate required field: author
	if strings.TrimSpace(prompt.Metadata.Author) == "" {
		return errors.ValidationError{
			Field:   "author",
			Message: "author is required and cannot be empty",
		}
	}

	// Validate author length (1-50 characters as per design doc)
	if len(prompt.Metadata.Author) > 50 {
		return errors.ValidationError{
			Field:   "author",
			Message: "author cannot be longer than 50 characters",
		}
	}

	// Validate optional fields if present
	if prompt.Metadata.Description != "" && len(prompt.Metadata.Description) > 500 {
		return errors.ValidationError{
			Field:   "description",
			Message: "description cannot be longer than 500 characters",
		}
	}

	// Validate tags if present
	for i, tag := range prompt.Metadata.Tags {
		if strings.TrimSpace(tag) == "" {
			return errors.ValidationError{
				Field:   "tags",
				Message: "tags cannot contain empty values",
			}
		}
		if len(tag) > 20 {
			return errors.ValidationError{
				Field:   "tags",
				Message: "each tag cannot be longer than 20 characters",
			}
		}
		// Check for duplicate tags
		for j := i + 1; j < len(prompt.Metadata.Tags); j++ {
			if tag == prompt.Metadata.Tags[j] {
				return errors.ValidationError{
					Field:   "tags",
					Message: "duplicate tags are not allowed",
				}
			}
		}
	}

	return nil
}
