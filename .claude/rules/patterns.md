# 通用模式规则

## 骨架项目

实现新功能时：
1. 搜索经过实战检验的骨架项目
2. 使用并行 Agent 评估选项（安全、扩展性、相关性、实现计划）
3. 克隆最佳匹配作为基础
4. 在成熟结构中迭代

## 设计模式

### Repository 模式
- 定义标准操作：findAll, findById, create, update, delete
- 具体实现处理存储细节
- 业务逻辑依赖抽象接口
- 便于测试和替换数据源

### API 响应格式
统一的响应封装：
- 包含 success/status 指示器
- 包含 data 载荷（错误时可为 null）
- 包含 error 消息字段
- 分页响应包含 metadata（total, page, limit）

### 不可变性模式
- 使用展开运算符创建新对象
- 数组操作使用 map/filter/reduce
- 避免直接修改传入的参数
- 状态更新总是返回新引用
