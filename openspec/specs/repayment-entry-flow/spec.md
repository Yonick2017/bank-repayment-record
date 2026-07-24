# repayment-entry-flow

## Purpose
定义还款录入流程、字段约束与提交后行为，确保录入交互一致且数据格式可验证。

## Requirements

### Requirement: Multi-step repayment entry flow
系统 MUST 提供三步式还款录入流程：步骤一选择银行卡、步骤二选择币种并输入金额、步骤三选择还款时间，并在步骤间提供页面过渡动画。

#### Scenario: Enter repayment through three steps
- **WHEN** 用户从首页进入“记录还款”并按顺序完成三个步骤
- **THEN** 系统按步骤推进并在每次点击“下一步”后展示统一过渡动画

### Requirement: Supported card options
系统 SHALL 在步骤一提供固定银行卡选项：`BOCHK Visa`、`BOCHK Mastercard`、`HSBC Visa Gold`、`HSBC Pulse`、`Hang Seng Travel+`、`HSBC Visa Signature`、`Amex US`、`BEA GOAL`、`CITIC Motion`、`Earnmore`、`SC Smart`、`ICBC SUP`、`ICBC 奋斗`，且用户 MUST 选择一项后方可进入下一步。

#### Scenario: Card selection is required
- **WHEN** 用户未选择银行卡就点击“下一步”
- **THEN** 系统阻止进入下一步并提示必须选择银行卡

### Requirement: Currency and amount input validation
系统 SHALL 在步骤二提供币种选项 `RMB` 与 `HKD`。金额输入 MUST 支持正数与负数，且精度为最多两位小数；负数在确认页与历史页 MUST 以 `CR` 形式展示。

#### Scenario: Negative amount shown as CR
- **WHEN** 用户输入负金额并完成录入
- **THEN** 系统保存该负值并在展示层以绝对值加 `CR` 显示

### Requirement: Time selection precision
系统 SHALL 在步骤三提供拨盘样式时间选择器，默认值为当前本地日期时间，用户可修改到分钟精度。

#### Scenario: Default time is current local minute
- **WHEN** 用户进入步骤三且未主动修改时间
- **THEN** 系统使用当前本地日期时间（分钟精度）作为默认提交值

### Requirement: Completion actions after submission
系统 SHALL 在提交成功后展示完成页，并提供 `再记一笔` 与 `查看历史` 两个操作入口。

#### Scenario: Completion page offers two actions
- **WHEN** 记录提交成功
- **THEN** 系统展示完成页并允许用户继续新增或跳转历史记录页面
