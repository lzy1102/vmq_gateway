"""
Python 接入示例 - V免签支付网关
支持回调模式和轮询模式
"""

import hashlib
import time
import requests
from flask import Flask, request, jsonify

# ========== 配置 ==========
GATEWAY_URL = "http://186.241.107.44"  # 网关地址
SERVICE_ID = "my_python_service"        # 你的服务ID
API_KEY = "xxxxxxxxxxxx"               # 绑定时生成的 API Key


# ========== 1. 创建订单 ==========
def create_order(amount: int, pay_type: str = "alipay", callback_url: str = ""):
    """
    创建支付订单

    Args:
        amount: 金额，单位分（100 = 1元）
        pay_type: wechat 或 alipay
        callback_url: 回调地址，不填则轮询模式

    Returns:
        dict: 包含订单信息
    """
    payload = {
        "amount": amount,
        "service_id": SERVICE_ID,
        "api_key": API_KEY,
        "pay_type": pay_type,
    }
    if callback_url:
        payload["callback_url"] = callback_url

    resp = requests.post(f"{GATEWAY_URL}/api/order", json=payload)
    data = resp.json()
    if data["code"] == 1:
        order = data["data"]
        print(f"订单创建成功: {order['order_id']}")
        print(f"  实付金额: {order['pay_str']} 元")
        print(f"  二维码: {order['qr_url']}")
        print(f"  有效期: {order['remaining_seconds']} 秒")
        return order
    else:
        print(f"创建失败: {data.get('msg', '未知错误')}")
        return None


# ========== 2. 查询订单状态（轮询模式）==========
def query_order(order_id: str):
    """
    查询订单支付状态

    Args:
        order_id: 订单号

    Returns:
        dict: 订单状态信息
    """
    resp = requests.get(f"{GATEWAY_URL}/api/order/status", params={"order_id": order_id})
    data = resp.json()
    if data["code"] == 1:
        return data["data"]
    return None


# ========== 3. 轮询示例 ==========
def poll_order(order_id: str, timeout: int = 900):
    """
    轮询等待订单支付完成

    Args:
        order_id: 订单号
        timeout: 超时时间（秒）

    Returns:
        str: paid / expired / timeout
    """
    start = time.time()
    print(f"开始轮询订单 {order_id}...")

    while time.time() - start < timeout:
        result = query_order(order_id)
        if result:
            status = result["status"]
            if status == "paid":
                print(f"✅ 收款成功！金额: {result['amount'] / 100} 元")
                return "paid"
            elif status in ("expired", "cancelled"):
                print(f"❌ 订单已{status}")
                return status
        time.sleep(3)

    print("⏰ 轮询超时")
    return "timeout"


# ========== 4. 回调模式 Flask 示例 ==========
app = Flask(__name__)


@app.route("/callback", methods=["POST"])
def payment_callback():
    """
    接收网关回调

    回调 POST body:
    {
        "trade_no": "V1783926073_103",
        "amount": 103,
        "pay_type": "alipay",
        "paid_at": 1783926200,
        "service_id": "my_service"
    }
    """
    data = request.json
    trade_no = data.get("trade_no")
    amount = data.get("amount", 0)
    pay_type = data.get("pay_type")

    print(f"收到回调: 订单 {trade_no}, 金额 {amount / 100} 元, 方式 {pay_type}")

    # TODO: 在这里处理业务逻辑
    # - 更新订单状态
    # - 发货/开通服务
    # - 记录日志

    return jsonify({"code": 1, "msg": "ok"})


# ========== 使用示例 ==========
if __name__ == "__main__":
    print("=" * 50)
    print("示例1: 创建订单（轮询模式）")
    print("=" * 50)

    # 创建 1 元订单，轮询模式
    order = create_order(amount=100, pay_type="alipay")
    if order:
        # 轮询等待支付
        poll_order(order["order_id"])

    print()
    print("=" * 50)
    print("示例2: 创建订单（回调模式）")
    print("=" * 50)

    # 创建 5 元订单，回调模式
    order = create_order(
        amount=500,
        pay_type="wechat",
        callback_url="https://your-server.com/callback"
    )

    # 启动 Flask 接收回调（生产环境用 gunicorn）
    # app.run(host="0.0.0.0", port=5000)
