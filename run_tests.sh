#!/bin/bash

echo "=============================="
echo "运行 OpenCode 项目单元测试"
echo "=============================="

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
total_modules=0
passed_modules=0
failed_modules=0

# 要测试的模块列表
modules=(
    "internal/config"
    "internal/fileutil"
    "internal/format"
    "internal/version"
    "internal/diff"
    "internal/permission"
    "internal/pubsub"
)

echo -e "${BLUE}开始运行单元测试...${NC}"
echo ""

# 运行每个模块的测试
for module in "${modules[@]}"; do
    echo -e "${YELLOW}测试模块: $module${NC}"
    ((total_modules++))
    
    # 运行测试，设置超时时间
    if go test ./$module -timeout=30s -v; then
        echo -e "${GREEN}✓ $module 测试通过${NC}"
        ((passed_modules++))
    else
        echo -e "${RED}✗ $module 测试失败${NC}"
        ((failed_modules++))
    fi
    echo ""
done

# 生成测试覆盖率报告
echo -e "${BLUE}生成测试覆盖率报告...${NC}"
go test ./internal/... -coverprofile=coverage.out -timeout=30s
if [ $? -eq 0 ]; then
    echo -e "${GREEN}覆盖率报告生成成功${NC}"
    go tool cover -func=coverage.out | grep "total:" || echo "总覆盖率信息不可用"
else
    echo -e "${YELLOW}覆盖率报告生成失败（部分测试可能有问题）${NC}"
fi

echo ""
echo "=============================="
echo "测试结果总结"
echo "=============================="
echo -e "总模块数: ${BLUE}$total_modules${NC}"
echo -e "通过模块: ${GREEN}$passed_modules${NC}"
echo -e "失败模块: ${RED}$failed_modules${NC}"

if [ $failed_modules -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试模块都通过了！${NC}"
    exit 0
else
    echo -e "${RED}⚠️  有 $failed_modules 个模块测试失败${NC}"
    exit 1
fi