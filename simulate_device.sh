#!/bin/bash
# 模拟V免签设备 - 持续发送心跳 + 模拟收款回调
# 用法: ./simulate_device.sh [服务器地址] [设备Key] [轮询间隔秒数]
#
# 示例:
#   ./simulate_device.sh http://186.241.107.44 f5b7e67ba6deb1bc23d3f1c39545f4ca
#   ./simulate_device.sh http://186.241.107.44 f5b7e67ba6deb1bc23d3f1c39545f4ca 30

BASE_URL="${1:-http://localhost}"
KEY="${2:-}"
INTERVAL="${3:-50}"

if [ -z "$KEY" ]; then
  echo "用法: $0 <服务器地址> <设备Key> [心跳间隔秒数]"
  echo "示例: $0 http://186.241.107.44 f5b7e67ba6deb1bc23d3f1c39545f4ca"
  exit 1
fi

green()  { echo -e "\033[32m$1\033[0m"; }
cyan()   { echo -e "\033[36m$1\033[0m"; }
yellow() { echo -e "\033[33m$1\033[0m"; }
red()    { echo -e "\033[31m$1\033[0m"; }

md5() {
  echo -n "$1" | md5sum | cut -d' ' -f1
}

send_heartbeat() {
  local t=$(($(date +%s%N) / 1000000))
  local sign=$(md5 "${t}${KEY}")
  local resp=$(curl -s "$BASE_URL/appHeart?t=${t}&sign=${sign}")
  if echo "$resp" | grep -q '"code":1'; then
    green "[心跳] $(date '+%H:%M:%S') 发送成功"
  else
    red "[心跳] $(date '+%H:%M:%S') 失败: $resp"
  fi
}

simulate_payment() {
  local pay_type="${1:-2}"
  local price="${2:-1.01}"
  local t=$(($(date +%s%N) / 1000000))
  local sign=$(md5 "${pay_type}${price}${t}${KEY}")
  local resp=$(curl -s "$BASE_URL/appPush?t=${t}&type=${pay_type}&price=${price}&sign=${sign}")
  if echo "$resp" | grep -q '"code":1'; then
    green "[收款] 支付成功: type=$pay_type price=$price"
  else
    yellow "[收款] 未匹配: $resp"
  fi
}

cyan "========================================="
cyan "  V免签设备模拟器"
cyan "  服务器: $BASE_URL"
cyan "  Key:    $KEY"
cyan "  心跳间隔: ${INTERVAL}s"
cyan "========================================="
cyan "  按 Ctrl+C 停止"
cyan "========================================="
echo ""

trap 'echo ""; yellow "设备模拟器已停止"; exit 0' INT TERM

while true; do
  send_heartbeat
  sleep "$INTERVAL"
done
