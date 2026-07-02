# local-storage-configuration

## Purpose
定义本地单用户数据持久化与数据库路径配置约束，确保服务在本地环境下稳定存取还款记录。

## Requirements

### Requirement: Local single-user persistence
系统 MUST 使用本地 SQLite 文件保存还款记录，并以单用户模式运行，不依赖登录或远程账号系统。

#### Scenario: Persist records locally
- **WHEN** 用户提交新的还款记录
- **THEN** 系统将记录写入本地 SQLite 并可在历史页读取

### Requirement: Configurable database file path
系统 SHALL 支持通过配置项设置 SQLite 文件路径；未配置时 MUST 使用默认路径。

#### Scenario: Use configured database path
- **WHEN** 系统启动且存在自定义数据库路径配置
- **THEN** 系统在该路径创建或打开数据库文件

### Requirement: Startup validation for database path
系统 MUST 在启动时校验数据库路径可访问与可写；若校验失败，系统 SHALL 返回明确错误并拒绝启动数据服务。

#### Scenario: Fail fast on invalid path
- **WHEN** 配置路径不存在且无法创建或无写权限
- **THEN** 系统启动失败并输出可读错误信息

### Requirement: Timestamp consistency
系统 SHALL 以统一时区规则保存还款时间，以确保按月分组与统计结果一致。

#### Scenario: Monthly grouping remains consistent
- **WHEN** 同一条记录在保存后被读取并参与月统计
- **THEN** 系统按统一时区归属到稳定的自然月分组
