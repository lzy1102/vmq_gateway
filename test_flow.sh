#!/bin/bash
# 完整流程测试 - 使用已有的设备和绑定
# 用法: ./test_flow.sh <服务器地址> <设备Key> <API Key> <服务ID> [金额分]
#
# 示例:
#   ./test_flow.sh http://186.241.107.44 455e8da9b6f332b16f3eae68f5f510f5 c6c9b929bed664372213d968cda7c23e test123

set -e

BASE_URL="${1:-http://localhost}"
DEVICE_KEY="${2:-}"
API_KEY="${3:-}"
SERVICE_ID="${4:-}"
AMOUNT="${5:-100}"

if [ -z "$DEVICE_KEY" ] || [ -z "$API_KEY" ] || [ -z "$SERVICE_ID" ]; then
  echo "用法: $0 <服务器地址> <设备Key> <API Key> <服务ID> [金额(分)]"
  echo ""
  echo "示例:"
  echo "  $0 http://186.241.107.44 <设备Key> <API Key> <服务ID>"
  echo "  $0 http://186.241.107.44 <设备Key> <API Key> <服务ID> 200"
  exit 1
fi

green()  { echo -e "\033[32m$1\033[0m"; }
red()    { echo -e "\033[31m$1\033[0m"; }
cyan()   { echo -e "\033[36m$1\033[0m"; }

md5() {
  echo -n "$1" | md5sum | cut -d' ' -f1
}

cyan "========================================="
cyan "  V免签完整流程测试"
cyan "  服务器: $BASE_URL"
cyan "  设备Key: ${DEVICE_KEY:0:8}..."
cyan "  API Key: ${API_KEY:0:8}..."
cyan "  服务ID: $SERVICE_ID"
cyan "  申请金额: ${AMOUNT} 分 ($(echo "scale=2; $AMOUNT/100" | bc) 元)"
cyan "========================================="
echo ""

# 1. 心跳上线
cyan ">>> 1. 发送心跳（设备上线）"
T=$(($(date +%s%N) / 1000000))
SIGN=$(md5 "${T}${DEVICE_KEY}")
RESP=$(curl -s "$BASE_URL/appHeart?t=${T}&sign=${SIGN}")
if echo "$RESP" | grep -q '"code":1'; then
  green "    心跳成功，设备已上线"
else
  red "    心跳失败: $RESP"
  exit 1
fi

sleep 1

# 2. 创建订单
cyan ">>> 2. 创建订单（申请 ${AMOUNT} 分）"
RESP=$(curl -s -X POST "$BASE_URL/api/order" \
  -H 'Content-Type: application/json' \
  -d "{\"amount\":$AMOUNT, \"service_id\":\"$SERVICE_ID\", \"api_key\":\"$API_KEY\"}")
echo "    $RESP"

ORDER_ID=$(echo "$RESP" | grep -o '"order_id":"[^"]*"' | cut -d'"' -f4)
PAY_AMOUNT=$(echo "$RESP" | grep -o '"pay_amount":[0-9]*' | cut -d':' -f2)
PAY_STR=$(echo "$RESP" | grep -o '"pay_str":"[^"]*"' | cut -d'"' -f4)
EXPIRE=$(echo "$RESP" | grep -o '"remaining_seconds":[0-9]*' | cut -d':' -f2)

if [ -z "$ORDER_ID" ]; then
  red "    创建订单失败"
  exit 1
fi
green "    订单号: $ORDER_ID"
green "    实际支付: $PAY_STR 元 (${PAY_AMOUNT} 分)"
green "    有效期: ${EXPIRE} 秒"

# 3. 查询订单状态（付款前）
cyan ">>> 3. 查询订单状态（付款前）"
RESP=$(curl -s "$BASE_URL/api/order/status?order_id=${ORDER_ID}")
STATUS=$(echo "$RESP" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
REMAINING=$(echo "$RESP" | grep -o '"remaining_seconds":[0-9]*' | cut -d':' -f2)
green "    状态: $STATUS, 剩余: ${REMAINING}秒"

# 4. 模拟付款（支付宝）
cyan ">>> 4. 模拟支付宝付款回调"
T=$(($(date +%s%N) / 1000000))
SIGN=$(md5 "2${PAY_STR}${T}${DEVICE_KEY}")
RESP=$(curl -s "$BASE_URL/appPush?t=${T}&type=2&price=${PAY_STR}&sign=${SIGN}")
echo "    $RESP"
if echo "$RESP" | grep -q '"code":1'; then
  green "    付款成功！订单已匹配"
else
  red "    付款失败"
  exit 1
fi

sleep 1

# 5. 查询订单状态（付款后）
cyan ">>> 5. 查询订单状态（付款后）"
RESP=$(curl -s "$BASE_URL/api/order/status?order_id=${ORDER_ID}")
STATUS=$(echo "$RESP" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
if [ "$STATUS" = "paid" ]; then
  green "    状态: paid ✓"
else
  red "    状态异常: $STATUS"
fi

echo ""
cyan "========================================="
green "  测试完成！"
cyan "========================================="
