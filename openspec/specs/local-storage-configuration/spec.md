# local-storage-configuration

## Purpose
定义单用户数据持久化与数据库连接配置约束，确保服务通过 YAML 配置连接 MySQL 并稳定存取还款记录。

## Requirements

### Requirement: Single-user MySQL persistence
系统 MUST 使用 MySQL 保存还款记录，并以单用户模式运行，不依赖远程多账号体系；访问门禁由共享密码会话提供（见 `shared-password-auth`）。

#### Scenario: Persist records in MySQL
- **WHEN** 已通过共享密码鉴权的用户提交新的还款记录
- **THEN** 系统将记录写入配置的 MySQL 数据库并可在历史页读取

### Requirement: YAML database connection configuration
系统 SHALL 通过 YAML 配置文件指定 MySQL 连接参数（host、port、user、password、database 等）；可通过 `CONFIG_PATH` 指定配置文件路径，未设置时 MUST 在约定路径查找配置文件。

#### Scenario: Use configured MySQL connection
- **WHEN** 系统启动且存在合法 YAML 配置
- **THEN** 系统使用该配置中的 MySQL 参数建立连接

### Requirement: Startup validation for database connectivity
系统 MUST 在启动时校验 MySQL 配置完整，并对数据库执行连通性检查（Ping）；若校验或连接失败，系统 SHALL 返回明确错误并拒绝启动数据服务。系统 MUST NOT 在启动时自动建表。

#### Scenario: Fail fast on invalid connection
- **WHEN** 配置缺失必填项、时区非法，或无法 Ping 通 MySQL
- **THEN** 系统启动失败并输出可读错误信息

### Requirement: Timestamp consistency
系统 SHALL 以统一时区规则保存还款时间，以确保按月分组与统计结果一致。

#### Scenario: Monthly grouping remains consistent
- **WHEN** 同一条记录在保存后被读取并参与月统计
- **THEN** 系统按统一时区归属到稳定的自然月分组
