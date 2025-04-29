# topnheap

该包的实现与 `github.com/emirpasic/gods/queues/priorityqueue@v1.18.1` 库相比，在执行效率和内存使用上有所优势，`topnheap` 提供了更快的操作速度和更低的延迟

## 性能对比

通过基准测试对 `topnheap` 和 `priorityqueue` 进行了比较，结果如下：

| 库名称               | 平均操作时间   | 内存使用        | 分配次数   |
|----------------------|----------------|-----------------|------------|
| `topnheap`           | 260.6 毫秒/次 | 160MB/次        | 1001 万次  |
| `priorityqueue`      | 1336.3 毫秒/次 | 160MB/次        | 1000 万次  |

- **执行时间**：`topnheap` 在执行时间上明显优于 `priorityqueue`，每次操作时间约为 `priorityqueue` 的五分之一。
- **内存使用**：两者在内存使用上的差异较小，但 `topnheap` 的内存分配稍多一些，这与堆元素的管理方式相关。
- **分配次数**：`topnheap` 的内存分配次数略高，但差异不大。

## 安装

```bash
go get github.com/nuominmin/biz/queues/topnheap
```

