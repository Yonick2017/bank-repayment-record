# repayment-history-and-stats

## Purpose
定义历史记录展示、筛选、统计与维护约束，确保用户可按一致口径查看并管理还款数据。

## Requirements

### Requirement: Monthly history listing
系统 SHALL 按月份分组展示还款记录，并在每个月分组下展示该月明细条目。

#### Scenario: Group records by month
- **WHEN** 用户进入“查看历史记录”页面
- **THEN** 系统按月份输出分组列表并显示每条记录的卡片信息

### Requirement: Currency-separated monthly statistics
系统 MUST 对 `RMB` 与 `HKD` 分开统计，不得跨币种合并求和或平均。

#### Scenario: Display separate totals by currency
- **WHEN** 页面计算月度统计
- **THEN** 系统分别输出 RMB 与 HKD 的统计结果

### Requirement: Average monthly repayment formula
系统 SHALL 提供“平均月开销”指标，其计算口径 MUST 为：同币种范围内，所有有记录月份的月总额之和除以有记录月份数（`sum(monthly_total) / months_with_records`）。

#### Scenario: Compute average across months with records
- **WHEN** 某币种在多个自然月存在记录
- **THEN** 系统按“月总额之和 / 有记录月数”输出该币种平均月开销

### Requirement: History filtering in P0
系统 SHALL 在历史页提供按银行卡与按币种筛选能力，筛选结果 MUST 同步影响明细与统计区域。

#### Scenario: Filters update list and stats together
- **WHEN** 用户选择银行卡或币种筛选条件
- **THEN** 系统仅展示匹配记录，并以同一筛选范围重算统计

### Requirement: Delete-only record maintenance
系统 SHALL 支持删除历史记录且 MUST 不提供编辑入口；删除操作 MUST 进行二次确认。

#### Scenario: Confirm before delete
- **WHEN** 用户点击删除某条记录
- **THEN** 系统弹出二次确认，确认后才删除并刷新列表
