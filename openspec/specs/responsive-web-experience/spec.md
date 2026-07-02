# responsive-web-experience

## Purpose
定义跨端布局与关键交互表现，确保移动端与桌面端在核心流程上具备一致可用性。

## Requirements

### Requirement: Responsive layouts for mobile and desktop
系统 SHALL 在手机端与桌面端提供可用且一致的核心流程，包括首页入口、录入步骤、完成页与历史页。

#### Scenario: Core pages adapt across viewport sizes
- **WHEN** 用户在手机或桌面浏览器访问应用
- **THEN** 系统根据视口调整布局且不丢失关键交互能力

### Requirement: Full-screen card selection page
系统 MUST 在录入步骤一使用全屏卡片式银行卡选择界面，保证触控可用性与视觉聚焦。

#### Scenario: Select card with touch-friendly UI
- **WHEN** 用户在移动设备进入步骤一
- **THEN** 系统展示全屏可点击卡片并支持单击选择

### Requirement: Consistent page transition animations
系统 SHALL 在录入各步骤与完成页之间使用统一的过渡动画，动画效果 MUST 不阻断用户操作且在低性能设备可平滑执行。

#### Scenario: Navigate with consistent transitions
- **WHEN** 用户在步骤间点击“下一步”或“完成”
- **THEN** 系统展示一致过渡动画并在动画后进入目标页面

### Requirement: Home summary card visibility
系统 SHALL 在首页展示“本月已还款总额”摘要卡片，并按 RMB/HKD 分列显示。

#### Scenario: Show current month totals on home
- **WHEN** 用户打开首页
- **THEN** 系统显示当月各币种还款总额摘要
