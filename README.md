# kt8900copilot-server

ESP32-S3 remote controller for QYT-KT8900 (Server Application)

Author: BG4QBF

esp32s3端见仓库 [odorajbotoj/kt8900copilot](https://github.com/odorajbotoj/kt8900copilot)

## 简介

客户端连接, 服务端发起验证, 验证通过后进行数据转发.

配置文件 `clients.json` 可以配置客户端的以下内容:

+ 客户端标识符
+ 客户端类型
+ 客户端名
+ 客户端 MAC 地址 ( 针对ESP32S3 )
+ 自己的输出连接到哪些客户端
+ 客户端验证密钥
+ 忽略特定数据包 ( 10进制整数 )

```json
{
    "BG4QBF": {
        "ClientId": "BG4QBF",
        "ClientType": 2,
        "ClientName": "BG4QBF",
        "ClientMac": "",
        "OutClientsNames": [
            "N0CALL"
        ],
        "Passkey": "key",
        "IgnoreFromChannel": [
            25,
            26,
            27
        ],
        "IgnoreFromWs": []
    }
}
```

服务器需要 **TLS** , 可以使用 *Let's Encrypt* 等进行证书管理.

## AIGC 内容告知

本仓库部分代码参考 *ChatGPT* 结果. 主要为 *HTML Client* 音频流那一部分.

## 免责声明

使用请遵守相关法律法规, 无线电相关内容请遵循无线电管理规定.

## License

MIT
