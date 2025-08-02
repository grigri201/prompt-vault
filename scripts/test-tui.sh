#!/bin/bash

# TUI 测试辅助脚本
# 用于运行需要 TTY 环境的集成测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查终端支持
check_tty_support() {
    print_info "检查 TTY 支持..."
    
    if [ ! -t 0 ]; then
        print_warning "标准输入不是 TTY，将设置虚拟终端环境"
        export TERM=${TERM:-xterm-256color}
        export FORCE_TTY=1
    fi
    
    if [ ! -t 1 ]; then
        print_warning "标准输出不是 TTY，将设置虚拟终端环境"
        export TERM=${TERM:-xterm-256color}
        export FORCE_TTY=1
    fi
    
    # 设置终端相关环境变量
    export TERM=${TERM:-xterm-256color}
    export COLUMNS=${COLUMNS:-80}
    export LINES=${LINES:-24}
    
    print_success "TTY 环境配置完成"
}

# 运行 TUI 单元测试
run_tui_unit_tests() {
    print_info "运行 TUI 组件单元测试..."
    
    # 设置测试环境变量
    export GO_TEST_ENV=tui
    export TUI_TEST_MODE=unit
    
    # 运行 TUI 相关的单元测试
    if go test -v ./internal/tui/... -tags=unit -coverprofile=coverage-tui-unit.out; then
        print_success "TUI 单元测试通过"
    else
        print_error "TUI 单元测试失败"
        return 1
    fi
}

# 运行 TUI 集成测试
run_tui_integration_tests() {
    print_info "运行 TUI 集成测试..."
    
    # 设置集成测试环境变量
    export GO_TEST_ENV=tui
    export TUI_TEST_MODE=integration
    export TEST_TIMEOUT=30s
    
    # 运行集成测试
    if go test -v ./integration/... -tags=tui -timeout=${TEST_TIMEOUT} -coverprofile=coverage-tui-integration.out; then
        print_success "TUI 集成测试通过"
    else
        print_error "TUI 集成测试失败"
        return 1
    fi
}

# 生成测试覆盖率报告
generate_coverage_report() {
    print_info "生成测试覆盖率报告..."
    
    # 合并覆盖率文件
    if command -v gocovmerge >/dev/null 2>&1; then
        gocovmerge coverage-tui-*.out > coverage-tui-total.out
        print_info "使用 gocovmerge 合并覆盖率文件"
    else
        print_warning "gocovmerge 未安装，使用第一个覆盖率文件"
        cp coverage-tui-unit.out coverage-tui-total.out 2>/dev/null || true
    fi
    
    # 生成 HTML 报告
    if [ -f coverage-tui-total.out ]; then
        go tool cover -html=coverage-tui-total.out -o coverage-tui.html
        print_success "覆盖率报告已生成: coverage-tui.html"
        
        # 显示覆盖率统计
        coverage_percent=$(go tool cover -func=coverage-tui-total.out | tail -1 | awk '{print $3}')
        print_info "TUI 测试覆盖率: ${coverage_percent}"
    fi
}

# 清理测试文件
cleanup() {
    print_info "清理临时文件..."
    rm -f coverage-tui-*.out
}

# 显示帮助信息
show_help() {
    cat << EOF
TUI 测试脚本

用法:
    $0 [选项]

选项:
    -u, --unit          只运行单元测试
    -i, --integration   只运行集成测试
    -c, --coverage      生成覆盖率报告
    -h, --help          显示帮助信息
    --clean             清理测试文件

示例:
    $0                  # 运行所有 TUI 测试
    $0 -u               # 只运行单元测试
    $0 -i -c            # 运行集成测试并生成覆盖率报告

环境变量:
    TERM                终端类型 (默认: xterm-256color)
    COLUMNS             终端列数 (默认: 80)
    LINES               终端行数 (默认: 24)
    TEST_TIMEOUT        测试超时时间 (默认: 30s)
    TUI_DEBUG           启用 TUI 调试输出

EOF
}

# 主函数
main() {
    local run_unit=true
    local run_integration=true
    local generate_coverage=false
    local clean_only=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--unit)
                run_unit=true
                run_integration=false
                shift
                ;;
            -i|--integration)
                run_unit=false
                run_integration=true
                shift
                ;;
            -c|--coverage)
                generate_coverage=true
                shift
                ;;
            --clean)
                clean_only=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 如果只需要清理，执行清理并退出
    if [ "$clean_only" = true ]; then
        cleanup
        exit 0
    fi
    
    print_info "开始 TUI 测试..."
    print_info "工作目录: $(pwd)"
    
    # 检查 Go 环境
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go 命令未找到，请确保 Go 已正确安装"
        exit 1
    fi
    
    # 检查项目结构
    if [ ! -f go.mod ]; then
        print_error "未找到 go.mod 文件，请在项目根目录运行此脚本"
        exit 1
    fi
    
    # 设置 TTY 环境
    check_tty_support
    
    # 执行测试
    local test_failed=false
    
    if [ "$run_unit" = true ]; then
        if ! run_tui_unit_tests; then
            test_failed=true
        fi
    fi
    
    if [ "$run_integration" = true ]; then
        if ! run_tui_integration_tests; then
            test_failed=true
        fi
    fi
    
    # 生成覆盖率报告
    if [ "$generate_coverage" = true ]; then
        generate_coverage_report
    fi
    
    # 检查测试结果
    if [ "$test_failed" = true ]; then
        print_error "部分测试失败"
        exit 1
    else
        print_success "所有 TUI 测试通过！"
    fi
}

# 设置错误处理
trap cleanup EXIT

# 执行主函数
main "$@"