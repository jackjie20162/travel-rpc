# 产品发布与行程信息优化

## 概述

本次优化参考携程产品发布页面，对 travel 项目的产品发布和商品管理进行全面升级，涵盖产品信息、行程信息、套餐报价、预订须知四大模块。重点优化了行程信息中的活动节点，新增 POI 搜索、入内/不入内、体验时长、活动特色、接送地点等功能。

## 优化范围

### 1. 产品信息增强
- 产品亮点（多条文本，JSON 存储）
- 封面图 + 多图上传
- 视频 URL
- 图文介绍（富文本 HTML）
- 预订须知文本

### 2. 行程信息重构（本次重点）
- 5 种节点类型：集合(MEETING)、活动(ACTIVITY)、交通(TRANSPORT)、餐饮(MEAL)、返程(RETURN)
- 活动节点特有字段：
  - 高德 POI 信息（poi_id, poi_name）
  - 入内/不入内选择
  - 体验时长模式（固定/按套餐/不限）
  - 活动体验特色描述
  - 接送地点（接站/送站的位置和坐标）
- 协议勾选：无购物承诺、行程可调整

### 3. 套餐与报价
- 报价模式：大小同价(SAME_PRICE)、区分人群(GROUP_PRICE)、阶梯报价(TIER_PRICE)
- 库存模式：无限库存(UNLIMITED)、日库存(DAILY)、总库存(TOTAL)
- 日历库存管理界面

### 4. 预订须知
- 退改规则配置
- 违约金设置

## 技术实现

### 后端变更

#### Ent Schema 变更

**文件**: `travel-rpc/ent/schema/itinerary_stop.go`

新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| poi_id | string | 高德 POI ID |
| poi_name | string | POI 名称 |
| is_entering | bool | 是否入内 |
| duration_mode | string | 时长模式：FIXED/PER_PACKAGE/UNLIMITED |
| duration_hours | int | 体验时长-小时 |
| duration_minutes | int | 体验时长-分钟 |
| activity_features | text | 活动体验特色 |
| pickup_location | string | 接站地点名称 |
| pickup_address | string | 接站地址 |
| pickup_latitude | float | 接站纬度 |
| pickup_longitude | float | 接站经度 |
| dropoff_location | string | 送站地点名称 |
| dropoff_address | string | 送站地址 |
| dropoff_latitude | float | 送站纬度 |
| dropoff_longitude | float | 送站经度 |

#### Protobuf 变更

**文件**: `travel-rpc/desc/travel.proto`

- `CreateItineraryStopRequest` 新增 14 个字段（field 18-31）
- `UpdateItineraryStopRequest` 新增 14 个字段（field 19-32）
- `ItineraryStop` message 新增 14 个字段（field 19-32）

#### RPC Service 变更

**文件**: `travel-rpc/internal/service/management.go`

- 更新 `itineraryStopMessage()` 函数映射所有新字段
- 更新 `CreateItineraryStop()` 方法处理新增字段
- 更新 `UpdateItineraryStop()` 方法处理新增字段（含 Clear 逻辑）

#### API Handler 变更

**文件**: `travel-api/internal/handler/types.go`

- `ItineraryStopReq` 新增 14 个可选字段
- `ItineraryStop` 新增 14 个字段

**文件**: `travel-api/internal/handler/management.go`

- 更新 `toItineraryStop()` 函数
- 更新 `CreateItineraryStop` handler
- 更新 `UpdateItineraryStop` handler

### 前端变更

**文件**: `merchant-frontend/src/plugin/travel/products/ProductEdit.vue`

#### UI 重构

Step 2（行程信息）完全重构，活动节点包含：

1. **基础信息区**
   - 节点类型选择
   - 标题输入

2. **活动详情区**（仅 ACTIVITY 类型显示）
   - POI 搜索框（带搜索按钮，显示 POI ID 标签）
   - 入内/不入内单选
   - 体验时长模式选择 + 时长输入
   - 活动特色多行文本框

3. **接送地点区**（仅 ACTIVITY 类型显示）
   - 接站地点名称 + 地址
   - 送站地点名称 + 地址

4. **通用信息区**
   - 开始/结束时间
   - 交通工具（TRANSPORT 类型）
   - 持续时间（TRANSPORT 类型）
   - 上门接/送返复选框

#### 数据结构

```javascript
{
  stopType: 'ACTIVITY',
  title: '',
  locationName: '',
  address: '',
  startTime: '',
  endTime: '',
  transportType: '',
  durationMinutes: 30,
  isPickup: false,
  isDropoff: false,
  // 新增字段
  poiId: '',
  poiName: '',
  isEntering: true,
  durationMode: 'FIXED',
  durationHours: 0,
  activityFeatures: '',
  pickupLocation: '',
  pickupAddress: '',
  pickupLatitude: 0,
  pickupLongitude: 0,
  dropoffLocation: '',
  dropoffAddress: '',
  dropoffLatitude: 0,
  dropoffLongitude: 0,
}
```

#### 新增函数

- `searchPoi(stop)` - POI 搜索（预留高德 API 接入）

## 编译验证

- ✅ travel-rpc 编译通过（`go build ./internal/...`）
- ✅ travel-api 编译通过（`go build ./...`）

## 待办事项

### 高优先级
- [ ] 接入高德地图 POI 搜索 API
- [ ] 执行数据库迁移（新增 ItineraryStop 表和扩展字段）
- [ ] 前端页面实际运行测试

### 中优先级
- [ ] 接送地点地图选点功能（高德地图组件）
- [ ] POI 搜索结果下拉选择
- [ ] 活动节点拖拽排序

### 低优先级
- [ ] 不同套餐时长不同的 UI 实现
- [ ] 行程节点预览模式
- [ ] 批量编辑行程节点

## 文件清单

### 后端文件
```
travel-rpc/
├── ent/schema/itinerary_stop.go      # 新增字段
├── desc/travel.proto                  # 新增字段
├── internal/service/management.go     # 更新方法
└── travel/                            # 生成的 pb.go 文件

travel-api/
├── internal/handler/types.go          # 新增类型字段
└── internal/handler/management.go     # 更新 handler
```

### 前端文件
```
merchant-frontend/
└── src/plugin/travel/
    ├── products/ProductEdit.vue       # 完全重构
    ├── packages/Index.vue             # 日历库存管理
    └── api.js                         # 新增 API 导出
```

## 注意事项

1. **数据库迁移**: 新增的 ItineraryStop 表和扩展字段需要执行 migration
2. **高德 API**: 当前 POI 搜索为模拟数据，需要配置高德 API Key
3. **字段命名**: Protobuf 下划线字段在 Go 中生成大驼峰命名（如 poi_id → PoiId）
4. **Ent 字段命名**: Ent 生成的字段名遵循 Go 命名规范（如 VideoURL 而非 VideoUrl）

## 参考

- 携程产品发布页面（10 张截图）
- 高德地图 POI 搜索 API 文档
- Element Plus 组件文档
