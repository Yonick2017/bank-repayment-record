# shared-password-auth

## Purpose
定义整站共享密码门禁：配置文件保存密码哈希，前端登录前哈希后再提交，后端以 HttpOnly Cookie 维持会话并保护业务 API。

## Requirements

### Requirement: Shared password gate without username
系统 MUST 使用整站单一共享密码作为访问门禁，不提供用户名；未通过鉴权的客户端 MUST NOT 访问业务页面所依赖的业务 API。

#### Scenario: Unauthenticated API access is rejected
- **WHEN** 客户端在未持有有效会话 Cookie 时请求业务 API（如还款或统计接口）
- **THEN** 系统返回 401，且不返回业务数据

### Requirement: Password hash stored in YAML
系统 MUST 在 YAML 配置的 `auth` 段保存登录密码的 SHA-256 十六进制摘要（`password_hash`）与会话签名密钥（`session_secret`）；启动时 MUST 校验二者合法，否则拒绝启动。

#### Scenario: Missing or invalid auth config fails startup
- **WHEN** 配置缺少 `auth.password_hash` / `auth.session_secret`，或 `password_hash` 不是 64 位十六进制
- **THEN** 系统启动失败并输出可读错误信息

### Requirement: Client-side hash before login request
登录请求体 MUST 携带密码的 SHA-256 十六进制摘要而非明文；后端 MUST 与配置中的 `password_hash` 做恒定时间比较。

#### Scenario: Correct hash issues session cookie
- **WHEN** 客户端提交与配置一致的 `passwordHash`
- **THEN** 系统签发名为 `brr_session` 的 HttpOnly Cookie（默认有效期 30 天，`SameSite=Lax`），后续业务请求可凭该 Cookie 通过鉴权

#### Scenario: Incorrect hash is rejected
- **WHEN** 客户端提交与配置不一致的 `passwordHash`
- **THEN** 系统返回错误且不签发有效业务会话 Cookie

### Requirement: Session check and logout
系统 MUST 提供会话状态查询与登出能力；登出后会话 Cookie MUST 失效，需重新登录方可访问业务 API。

#### Scenario: Me endpoint reports session
- **WHEN** 客户端携带有效会话 Cookie 请求 `GET /api/auth/me`
- **THEN** 系统返回成功状态

#### Scenario: Logout clears session
- **WHEN** 客户端调用登出接口
- **THEN** 会话 Cookie 被清除，随后业务 API 请求返回 401
