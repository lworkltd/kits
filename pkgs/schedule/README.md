schedule

-----------------
调度包，多种调度策略

功能点
-----
### 调度器
规定了什么逻辑条件下，什么时间条件下执行指定任务
1.定期执行任务  
```
TODO:EXAMPLES
```
2.间隔固定时间执行任务  
```
TODO:EXAMPLES
```
3.条件满足时执行任务  
```
TODO:EXAMPLES
```
4.条件满足时退出   
```
TODO:EXAMPLES
```
5.安全执行调度  
```
TODO:EXAMPLES
```
6.延迟调度  
```
TODO:EXAMPLES
```

### 例子
##### 简单入门
```
got := New().
Safety(). // 拦截panic
Every(time.Millisecond * 50). // 执行间隔
Start(func() { // 开始执行任务
    fmt.Printf("Hello World")
})
```

##### 更多功能 
```
c := 0
// 执行条件
cond := func() bool {
    c++
    return c < 3 || c > 5
}
// 停止条件
closeCond := func() bool {
    return c == 10
}
wg.Add(7)
got := New().
    If(cond). // 执行的条件
    Count(10). // 有效执行最多10次
    Safety(). // 拦截panic
    CloseIf(closeCond). // 停止条件
    Delay(time.Millisecond * 20). // 延迟启动
    Every(time.Millisecond * 50). // 执行间隔
    Start(func() { // 开始执行任务
    fmt.Printf("########### %d\n", c)
    wg.Done()
})
wg.Wait()
```

最新
----
也许这个库有点多余