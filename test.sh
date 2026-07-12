#!/bin/bash
set -e

BASE_URL="${1:-http://localhost}"
KEY=""
DEVICE_ID="test_device_$(date +%s)"
COOKIE_FILE=$(mktemp)

green() { echo -e "\033[32m$1\033[0m"; }
red()   { echo -e "\033[31m$1\033[0m"; }
info()  { echo -e "\033[36m$1\033[0m"; }

md5() {
  echo -n "$1" | md5sum | cut -d' ' -f1
}

cleanup() { rm -f "$COOKIE_FILE"; }
trap cleanup EXIT

info "=== 0. 登录 ==="
RESP=$(curl -s -X POST "$BASE_URL/admin/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"vmq_gateway"}' \
  -c "$COOKIE_FILE")
echo "$RESP"
if ! echo "$RESP" | grep -q '"code":1'; then
  red "登录失败"
  exit 1
fi
green "登录成功"

info "=== 1. 创建设备 ==="
RESP=$(curl -s -X POST "$BASE_URL/admin/device" \
  -H 'Content-Type: application/json' \
  -b "$COOKIE_FILE" \
  -d "{\"device_id\":\"$DEVICE_ID\"}")
echo "$RESP"
KEY=$(echo "$RESP" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
if [ -z "$KEY" ]; then
  red "创建设备失败"
  exit 1
fi
green "设备ID: $DEVICE_ID"
green "Key: $KEY"

info "=== 1.5 创建服务绑定 ==="
RESP=$(curl -s -X POST "$BASE_URL/admin/binding" \
  -H 'Content-Type: application/json' \
  -b "$COOKIE_FILE" \
  -d "{\"service_id\":\"test_service\",\"callback_url\":\"https://httpbin.org/post\"}")
echo "$RESP"
API_KEY=$(echo "$RESP" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
if [ -z "$API_KEY" ]; then
  red "创建绑定失败"
  exit 1
fi
green "API Key: $API_KEY"

info "=== 2. 模拟心跳 ==="
T=$(($(date +%s%N) / 1000000))
SIGN=$(md5 "${T}${KEY}")
RESP=$(curl -s "$BASE_URL/appHeart?t=${T}&sign=${SIGN}")
echo "$RESP"
if echo "$RESP" | grep -q '"code":1'; then
  green "心跳成功"
else
  red "心跳失败"
fi

sleep 1

info "=== 3. 创建订单 ==="
RESP=$(curl -s -X POST "$BASE_URL/api/order" \
  -H 'Content-Type: application/json' \
  -d "{\"amount\":100, \"service_id\":\"test_service\", \"callback_url\":\"https://httpbin.org/post\", \"api_key\":\"$API_KEY\"}")
echo "$RESP"
ORDER_ID=$(echo "$RESP" | grep -o '"order_id":"[^"]*"' | cut -d'"' -f4)
AMOUNT_YUAN=$(echo "$RESP" | grep -o '"amount_str":[0-9.]*' | cut -d':' -f2)

if [ -z "$ORDER_ID" ]; then
  red "创建订单失败"
  exit 1
fi
green "订单号: $ORDER_ID"
green "支付金额: ${AMOUNT_YUAN} 元"

info "=== 4. 查询订单状态（付款前）==="
RESP=$(curl -s "$BASE_URL/api/order/status?order_id=${ORDER_ID}")
echo "$RESP"

info "=== 5. 模拟付款回调 ==="
T=$(($(date +%s%N) / 1000000))
SIGN=$(md5 "2${AMOUNT_YUAN}${T}${KEY}")
RESP=$(curl -s "$BASE_URL/appPush?t=${T}&type=2&price=${AMOUNT_YUAN}&sign=${SIGN}")
echo "$RESP"
if echo "$RESP" | grep -q '"code":1'; then
  green "付款回调成功！订单已匹配"
else
  red "付款回调失败"
fi

sleep 1

info "=== 6. 查询订单状态（付款后）==="
RESP=$(curl -s "$BASE_URL/api/order/status?order_id=${ORDER_ID}")
echo "$RESP"
if echo "$RESP" | grep -q '"status":"paid"'; then
  green "订单状态已变为 paid ✓"
else
  red "订单状态异常"
fi

info "=== 7. 检查回调通知 ==="
info "httpbin.org/post 应该收到了 POST 回调，内容包含:"
green '{"order_id":"'"$ORDER_ID"'", "amount":..., "service_id":"test_service", "status":"paid"}'

echo ""
green "=== 测试完成 ==="
