# hashtable
[![Go Report Card](https://goreportcard.com/badge/github.com/yeqown/hashtable)](https://goreportcard.com/report/github.com/yeqown/hashtable) [![](https://godoc.org/github.com/yeqown/hashtable?status.svg)](https://godoc.org/github.com/yeqown/hashtable)

自定义hash表实现，并发安全。

### TODOs:

[ ] 利用go:generate来生成指定Map<KeyTyp,ValueTyp>并生成方法集

[ ] 利用murmur3来替换murmur3算法

[x] 参照redis.dict实现链式hash表 http://zhangtielei.com/posts/blog-redis-dict.html

[x] 能自动扩容，rehash

[ ] 完成性能测试，对比golang内置Map



### Redis-Hashtable 特点

* 使用链地址法来解决键冲突
* 自动扩容
* 渐进式rehash