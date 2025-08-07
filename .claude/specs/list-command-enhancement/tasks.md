# 任务文档 - List 命令导出信息显示增强

## 任务概览

本文档将 `pv list` 命令中导出信息显示的实现分解为原子性、可执行的任务。每个任务都设计为可以独立完成并进行彻底测试。

## 实现任务

### 第一阶段：核心导出状态显示

#### 任务 1：添加导出数据结构
**文件**：`cmd/list.go`
**需求**：需求 1（导出状态显示增强）
**现有代码**：利用现有的 `model.IndexedPrompt` 和 Store 接口

- [ ] 将 `ExportStatus` 结构体添加到 `cmd/list.go`
  ```go
  type ExportStatus struct {
      IsExported bool
      ExportURL  string
      ExportedBy string
  }
  ```
- [ ] 确保结构体遵循现有的 Go 命名约定
- [ ] 按照项目标准添加结构体文档
- [ ] 验证与现有类型没有命名冲突

**验收标准**：导出状态数据结构得到正确定义和记录

---

#### 任务 2：实现导出映射构建器
**文件**：`cmd/list.go`
**需求**：需求 1（导出状态显示增强）
**现有代码**：使用现有的 `store.GetExports()` 方法和 `model.IndexedPrompt`

- [ ] Implement `buildExportMap(exports []model.IndexedPrompt) map[string]ExportStatus` function
- [ ] Handle nil `Parent` field gracefully
- [ ] Handle empty `Parent` string gracefully
- [ ] Use `export.GistURL` as the export URL value
- [ ] Map `export.Parent` (original prompt URL) to export status
- [ ] Add comprehensive function documentation

**验收标准**：导出映射正确关联原始 prompt 与其导出数据

---

#### 任务 3：创建导出信息格式化器
**Files**: `cmd/list.go`
**Requirements**: Requirement 1 (Export Status Display Enhancement), AC-1, AC-3
**Existing Code**: Follow existing formatting patterns in list command

- [ ] Implement `formatExportInfo(status ExportStatus) string` function
- [ ] Return `[not exported]` for non-exported prompts
- [ ] Return `[export status unknown]` for invalid export URLs
- [ ] Return `[✓ exported: <url>]` for valid exports
- [ ] Ensure Unicode checkmark works or fallback to ASCII
- [ ] Add proper error handling for edge cases

**验收标准**：导出信息格式一致并处理所有边界情况

---

#### 任务 4：实现增强的 Prompt 显示
**Files**: `cmd/list.go`  
**Requirements**: Requirement 1 (Export Status Display Enhancement), AC-1, AC-2
**Existing Code**: Extend existing prompt formatting logic

- [ ] Implement `formatPromptWithExport(prompt model.Prompt, exportMap map[string]ExportStatus) string`
- [ ] Lookup export status using `prompt.GistURL` as key
- [ ] Preserve existing format: `"  %s - author: %s : %s"`
- [ ] Append export information: `" %s"`
- [ ] Handle missing export data gracefully
- [ ] Maintain consistent spacing and formatting

**验收标准**：增强显示包含导出信息并保持现有格式

---

#### Task 5: Integrate Export Data Retrieval
**Files**: `cmd/list.go`
**Requirements**: Requirement 1 (Export Status Display Enhancement), AC-5  
**Existing Code**: Modify existing `(*list).execute()` method

- [ ] Add export data retrieval after `store.List()` call
- [ ] Call `store.GetExports()` to get export data
- [ ] Handle `GetExports()` error gracefully (continue without export info)
- [ ] Build export map using `buildExportMap()` function  
- [ ] Ensure error handling doesn't break existing functionality
- [ ] Maintain existing error patterns and logging

**Acceptance Criteria**: Export data is retrieved and integrated without affecting core list functionality

---

### Phase 2: Display Integration

#### Task 6: Update Display Loop
**Files**: `cmd/list.go`
**Requirements**: Requirement 1 (Export Status Display Enhancement), AC-1, AC-2, AC-3
**Existing Code**: Modify existing prompt display loop in `(*list).execute()`

- [ ] Replace `fmt.Printf("  %s - author: %s : %s\n ", ...)` with enhanced formatting
- [ ] Use `formatPromptWithExport()` for each prompt
- [ ] Pass `exportMap` to formatting function
- [ ] Ensure newline formatting remains consistent
- [ ] Preserve existing loop structure and variable names
- [ ] Maintain exact spacing and alignment

**Acceptance Criteria**: All prompts display with export information in consistent format

---

### Phase 3: Error Handling and Edge Cases

#### Task 7: Add Export Data Error Handling
**Files**: `cmd/list.go`
**Requirements**: Requirement 1 (Export Status Display Enhancement)
**Existing Code**: Use existing error handling patterns from list command

- [ ] Handle `store.GetExports()` returning error
- [ ] Create empty export map on error (graceful degradation)
- [ ] Log export retrieval errors at appropriate level
- [ ] Ensure list command continues to function with export errors
- [ ] Follow existing error logging patterns
- [ ] Add appropriate error messages for debugging

**Acceptance Criteria**: Export data errors don't break list functionality and are properly logged

---

#### Task 8: Handle Invalid Export URLs
**Files**: `cmd/list.go`  
**Requirements**: Security requirement for URL validation
**Existing Code**: Create new validation following existing patterns

- [ ] Add basic GitHub Gist URL validation in `formatExportInfo()`
- [ ] Check URL format: `https://gist.github.com/...`
- [ ] Handle empty or malformed URLs gracefully
- [ ] Return appropriate status message for invalid URLs
- [ ] Prevent potential security issues with malformed URLs
- [ ] Add URL validation documentation

**Acceptance Criteria**: Invalid export URLs are handled safely and display appropriate messages

---

### Phase 4: Backward Compatibility

#### Task 9: Ensure Backward Compatibility  
**Files**: `cmd/list.go`
**Requirements**: Requirement 2 (Backward Compatibility Preservation), AC-4
**Existing Code**: Preserve all existing interfaces and behavior

- [ ] Verify all existing command-line arguments work unchanged
- [ ] Ensure `--remote` flag behavior is preserved
- [ ] Maintain existing error message formats  
- [ ] Preserve cache timestamp display logic
- [ ] Verify empty prompt collection messages unchanged
- [ ] Test first-time user messages remain the same

**Acceptance Criteria**: All existing list command functionality works exactly as before

---

#### Task 10: Cache Integration Verification
**Files**: `cmd/list.go`
**Requirements**: Requirement 3 (Cache Integration Support), AC-5
**Existing Code**: Work with existing cache mechanism

- [ ] Verify `store.GetExports()` works with cached store
- [ ] Ensure `--remote` flag fetches fresh export data
- [ ] Verify cache timestamp display still works
- [ ] Test cache fallback behavior with export data
- [ ] Ensure cache manager error handling is preserved
- [ ] Verify cache-first behavior includes export information

**Acceptance Criteria**: Export information integrates seamlessly with existing cache mechanism

---

### Phase 5: Testing and Validation

#### Task 11: Add Unit Tests for Export Functions
**Files**: `cmd/list_test.go`
**Requirements**: All requirements (comprehensive testing)
**Existing Code**: Extend existing test patterns

- [ ] Test `buildExportMap()` with various input scenarios
- [ ] Test `formatExportInfo()` with all status types
- [ ] Test `formatPromptWithExport()` with and without export data
- [ ] Test edge cases: nil pointers, empty strings, malformed data
- [ ] Follow existing test naming and structure patterns
- [ ] Use existing mock patterns where applicable

**Acceptance Criteria**: All new functions have comprehensive unit test coverage

---

#### Task 12: Add Integration Tests
**Files**: `cmd/list_test.go`  
**Requirements**: All requirements (end-to-end validation)
**Existing Code**: Extend existing integration test patterns

- [ ] Test complete list command with export data
- [ ] Test list command with missing export data
- [ ] Test `--remote` flag with export information
- [ ] Test cache behavior with export data
- [ ] Test error scenarios and graceful degradation
- [ ] Use existing test infrastructure and mocks

**Acceptance Criteria**: Integration tests verify complete export display functionality

---

## Implementation Notes

### Dependencies Between Tasks
- Tasks 1-4 can be completed in parallel
- Task 5 depends on Tasks 1-2  
- Task 6 depends on Tasks 1-4
- Tasks 7-8 depend on Task 5
- Tasks 9-10 can be done in parallel after core tasks
- Testing tasks (11-12) should be done continuously with implementation

### Testing Strategy
- Each task should include basic testing during implementation
- Use existing mock infrastructure for Store interface
- Test both success and failure scenarios for each function
- Validate backward compatibility after each change

### Performance Considerations  
- `buildExportMap()` creates single lookup map for efficiency
- Export data retrieval adds one additional Store call
- Map lookup is O(1) for each prompt display
- Graceful degradation ensures no performance penalty on errors

### Risk Mitigation
- All existing functionality preserved through careful integration
- Error handling ensures export failures don't break core features  
- Comprehensive testing validates both new and existing behavior
- Incremental implementation allows for rollback if needed