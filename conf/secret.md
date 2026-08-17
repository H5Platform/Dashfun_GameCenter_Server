# `secret.yml` 配置说明

服务启动时会读取 `conf/secret.yml`，并使用其中的键替换 `conf/config.yml` 内对应的 `${...}` 占位符。

`conf/secret.yml` 已加入 `.gitignore`，不得提交到 Git。每个部署环境都需要单独创建该文件，并通过安全渠道保存和分发密钥。

## 配置模板

在 `conf/secret.yml` 中写入以下内容。尖括号内仅为说明性占位符，不是真实值，部署时必须替换。

```yaml
telegram_bot_token: <Telegram Bot Token>

aws_pinpoint_access_key_id: <AWS Access Key ID>
aws_pinpoint_access_key: <AWS Secret Access Key>

admin_password: <后台管理员初始密码>
admin_search_password: <后台数据查询密码>

tencent_cos_secret_id: <Tencent Cloud SecretId>
tencent_cos_secret_key: <Tencent Cloud SecretKey>

web3_op_private_key: <HHA Web3 运营钱包私钥>
squad_game_admin_key: <SquadGame 管理钱包私钥>

open_api_key: <OpenAI API Key>

account_token_secret: <本地账号 Token 签名密钥>
```

## 配置项来源

### `telegram_bot_token`

用于 Telegram Bot API、Telegram 文件访问及 Telegram 用户相关逻辑。

获取方式：在 Telegram 中联系 [BotFather](https://t.me/BotFather)，创建或选择对应 Bot 后获取 Token。开发、测试和生产环境应使用各自的 Bot。

### `aws_pinpoint_access_key_id` / `aws_pinpoint_access_key`

用于通过 AWS Pinpoint Email 发送注册验证码、密码重置验证码及后台激活邮件。

获取方式：在 AWS IAM 中为服务创建专用身份并生成 Access Key。该身份需要具备项目所用 Pinpoint Email 发送操作的最小权限。不要使用个人或根账号密钥。

### `admin_password`

用于初始化后台管理员账号。

获取方式：由部署负责人生成并存入团队使用的密码管理器。应使用独立的高强度随机密码，不得与个人密码或其他环境共用。

### `admin_search_password`

用于保护后台数据查询接口。

获取方式：由部署负责人单独生成并存入密码管理器。不得与 `admin_password` 使用相同内容。

### `tencent_cos_secret_id` / `tencent_cos_secret_key`

用于访问 Tencent COS 对象存储。

获取方式：在腾讯云访问管理（CAM）中创建项目专用子用户或角色并生成 API 密钥，仅授予目标 COS Bucket 所需的最小权限。

### `web3_op_private_key`

HHA Web3 运营钱包私钥，用于服务端发起相应链上操作。

获取方式：从项目指定的运营钱包安全导出。钱包地址、网络和资金必须与 `conf/config.yml` 中的 Web3 配置匹配。私钥应通过密钥管理系统或密码管理器传递，禁止通过聊天或代码仓库发送。

### `squad_game_admin_key`

SquadGame 管理钱包私钥，用于服务端执行 SquadGame 合约管理操作。

获取方式：从拥有对应合约管理权限的钱包安全导出，并确认它与 `hourly_squad_game_cfg.contract_address` 所指向合约的权限配置一致。建议与普通运营钱包分离。

### `open_api_key`

用于服务端调用 OpenAI API。

获取方式：在 [OpenAI Platform](https://platform.openai.com/api-keys) 中为对应项目创建 API Key。不同环境建议使用不同的 Project 和 Key，并设置合适的权限及用量限制。

### `account_token_secret`

用于本地账号系统签发和校验登录 Token。

获取方式：由部署负责人生成独立的高强度随机密钥。建议使用至少 256-bit 的安全随机值，并存入密码管理器。各环境应使用不同密钥；更换该密钥会使已有登录 Token 失效。

## 安全要求

- 不要在本文件、提交记录、日志、Issue 或聊天中填写真实值。
- 不要提交 `conf/secret.yml`，提交前应检查 `git status`。
- 开发、测试和生产环境必须使用不同密钥。
- 权限应遵循最小权限原则，并按团队安全策略定期轮换。
- 如果密钥曾出现在 Git 历史、日志或公开聊天中，应立即在对应平台撤销并重新生成，仅删除文本并不能消除泄露风险。
